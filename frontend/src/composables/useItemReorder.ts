import { nextTick, onUnmounted, ref, type Ref } from 'vue'
import type { NavGroup, NavItem } from '../api/settings'

const LONG_PRESS_MS = 360
const MOUSE_DRAG_START_PX = 16
const TOUCH_MOVE_CANCEL_PX = 14
const GROUP_MAGNET_PX = 20
const GROUP_INTENT_MS = 100
const FAST_POINTER_PX_MS = 1.4
const GAP_EXPAND_PX = 28
const ICON_HIT_PAD_PX = 6
const PLACEHOLDER_KEEP_PAD_PX = 10
const MOVE_MS = 180
const SETTLE_MS = 160
const RETURN_MS = 180
const AUTO_SCROLL_ZONE = 64
const ROW_CLUSTER_PX = 24

export type DragGhost = {
  item: NavItem
  x: number
  y: number
  width: number
  height: number
  returning: boolean
}

type Point = { x: number; y: number }

type ItemBox = {
  id: string
  index: number
  rect: DOMRect
}

function cloneGroups(groups: NavGroup[]): NavGroup[] {
  return groups.map((group) => ({
    ...group,
    items: (group.items || []).map((item) => ({ ...item })),
  }))
}

function sameOrder(a: NavGroup[], b: NavGroup[]) {
  if (a.length !== b.length) return false
  for (let i = 0; i < a.length; i += 1) {
    if (a[i].id !== b[i].id) return false
    const ai = a[i].items || []
    const bi = b[i].items || []
    if (ai.length !== bi.length) return false
    for (let j = 0; j < ai.length; j += 1) {
      if (ai[j].id !== bi[j].id) return false
    }
  }
  return true
}

function locateItem(groups: NavGroup[], itemId: string) {
  for (const group of groups) {
    const items = group.items || []
    const index = items.findIndex((item) => item.id === itemId)
    if (index >= 0) return { group, index }
  }
  return null
}

function prefersReducedMotion() {
  return window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false
}

function captureRects() {
  const map = new Map<string, DOMRect>()
  document.querySelectorAll<HTMLElement>('.home-board [data-item-id]').forEach((el) => {
    const id = el.dataset.itemId
    if (id) map.set(id, el.getBoundingClientRect())
  })
  return map
}

function playFlip(prev: Map<string, DOMRect>) {
  if (prefersReducedMotion() || prev.size === 0) return
  document.querySelectorAll<HTMLElement>('.home-board [data-item-id]').forEach((el) => {
    const id = el.dataset.itemId
    if (!id) return
    const first = prev.get(id)
    if (!first) return
    const last = el.getBoundingClientRect()
    const dx = first.left - last.left
    const dy = first.top - last.top
    if (Math.abs(dx) < 0.5 && Math.abs(dy) < 0.5) return
    el.style.transition = 'none'
    el.style.transform = `translate(${dx}px, ${dy}px)`
    void el.offsetWidth
    el.style.transition = `transform ${MOVE_MS}ms ease-out`
    el.style.transform = ''
    const cleanup = (event: TransitionEvent) => {
      if (event.propertyName && event.propertyName !== 'transform') return
      el.style.transition = ''
      el.style.transform = ''
      el.removeEventListener('transitionend', cleanup)
    }
    el.addEventListener('transitionend', cleanup)
  })
}

function inflateContains(rect: DOMRect, x: number, y: number, pad: number) {
  return (
    x >= rect.left - pad &&
    x <= rect.right + pad &&
    y >= rect.top - pad &&
    y <= rect.bottom + pad
  )
}

function distanceToRect(x: number, y: number, rect: DOMRect) {
  const dx = x < rect.left ? rect.left - x : x > rect.right ? x - rect.right : 0
  const dy = y < rect.top ? rect.top - y : y > rect.bottom ? y - rect.bottom : 0
  return Math.hypot(dx, dy)
}

function clusterRows(boxes: ItemBox[]) {
  const sorted = [...boxes].sort(
    (a, b) => a.rect.top - b.rect.top || a.rect.left - b.rect.left,
  )
  const rows: ItemBox[][] = []
  for (const box of sorted) {
    const row = rows.find((items) => Math.abs(items[0].rect.top - box.rect.top) < ROW_CLUSTER_PX)
    if (row) row.push(box)
    else rows.push([box])
  }
  for (const row of rows) row.sort((a, b) => a.rect.left - b.rect.left)
  return rows
}

function nearestRow(rows: ItemBox[][], y: number) {
  let best = rows[0]
  let bestDist = Infinity
  for (const row of rows) {
    const top = Math.min(...row.map((box) => box.rect.top))
    const bottom = Math.max(...row.map((box) => box.rect.bottom))
    const dist = y < top ? top - y : y > bottom ? y - bottom : 0
    if (dist < bestDist) {
      bestDist = dist
      best = row
    }
  }
  return best
}

function insertIndexAmong(others: ItemBox[], x: number, y: number) {
  if (!others.length) return 0
  const rows = clusterRows(others)
  const row = nearestRow(rows, y)

  for (const box of row) {
    if (inflateContains(box.rect, x, y, ICON_HIT_PAD_PX)) return box.index
  }

  if (x < row[0].rect.left + GAP_EXPAND_PX) return row[0].index

  for (let i = 0; i < row.length - 1; i += 1) {
    const left = row[i]
    const right = row[i + 1]
    const zoneLeft = left.rect.right - GAP_EXPAND_PX
    const zoneRight = right.rect.left + GAP_EXPAND_PX
    if (x >= zoneLeft && x <= zoneRight) return right.index
  }

  return row[row.length - 1].index + 1
}

function edgeScrollDelta(pos: number, size: number) {
  if (pos < AUTO_SCROLL_ZONE) {
    const t = 1 - Math.max(pos, 0) / AUTO_SCROLL_ZONE
    return -(4 + t * 24)
  }
  if (pos > size - AUTO_SCROLL_ZONE) {
    const t = 1 - Math.max(size - pos, 0) / AUTO_SCROLL_ZONE
    return 4 + t * 24
  }
  return 0
}

function itemEl(itemId: string) {
  return document.querySelector<HTMLElement>(`.home-board [data-item-id="${itemId}"]`)
}

export function useItemReorder(options: {
  groups: Ref<NavGroup[]>
  enabled: () => boolean
  onDragStart?: () => void
  onCommit: (payload: { destGroupId: string; destIds: string[] }) => Promise<void>
  onCommitError: () => Promise<void> | void
}) {
  const dragging = ref(false)
  const dragItemId = ref('')
  const dropGroupId = ref('')
  const hoverGroupId = ref('')
  const ghost = ref<DragGhost | null>(null)
  const suppressClick = ref(false)

  let snapshot: NavGroup[] = []
  let originGroupId = ''
  let pointerId = -1
  let pointerKind = ''
  let pointer: Point = { x: 0, y: 0 }
  let originPointer: Point = { x: 0, y: 0 }
  let lastPointer: Point = { x: 0, y: 0 }
  let lastPointerAt = 0
  let velocity = 0
  let grabOffset = { x: 0, y: 0 }
  let longPressTimer = 0
  let intentTimer = 0
  let pendingGroupId = ''
  let pendingItem: NavItem | null = null
  let originEl: HTMLElement | null = null
  let previewJob: { groupId: string; index: number } | null = null
  let raf = 0
  let sessionActive = false
  let finishing = false
  let saveQueue: Promise<void> = Promise.resolve()

  function isTouchLike() {
    return pointerKind === 'touch' || pointerKind === 'pen'
  }

  function releaseCapture() {
    if (!originEl || pointerId < 0) return
    try {
      if (originEl.hasPointerCapture(pointerId)) originEl.releasePointerCapture(pointerId)
    } catch {
      /* already released */
    }
  }

  function clearTimers() {
    window.clearTimeout(longPressTimer)
    window.clearTimeout(intentTimer)
    longPressTimer = 0
    intentTimer = 0
    pendingGroupId = ''
  }

  function stopLoop() {
    if (!raf) return
    cancelAnimationFrame(raf)
    raf = 0
  }

  function unbindWindow() {
    window.removeEventListener('pointermove', onPointerMove)
    window.removeEventListener('pointerup', onPointerUp)
    window.removeEventListener('pointercancel', onPointerUp)
    window.removeEventListener('keydown', onKeydown)
    window.removeEventListener('contextmenu', onContextMenu, true)
  }

  function resetSession() {
    clearTimers()
    stopLoop()
    unbindWindow()
    releaseCapture()
    document.body.classList.remove('is-item-dragging')
    sessionActive = false
    finishing = false
    pointerId = -1
    pointerKind = ''
    velocity = 0
    dragging.value = false
    dragItemId.value = ''
    dropGroupId.value = ''
    hoverGroupId.value = ''
    ghost.value = null
    snapshot = []
    originGroupId = ''
    pendingItem = null
    originEl = null
    previewJob = null
  }

  function beginDrag() {
    if (!sessionActive || dragging.value || finishing) return
    if (!pendingItem || !originEl) return
    window.clearTimeout(longPressTimer)
    longPressTimer = 0
    startDrag(pendingItem, originEl)
  }

  function hitTestGroup(x: number, y: number) {
    const sections = [...document.querySelectorAll<HTMLElement>('.home-board [data-group-id]')]
    let best: { id: string; dist: number; inside: boolean } | null = null
    for (const section of sections) {
      const id = section.dataset.groupId
      if (!id) continue
      const rect = section.getBoundingClientRect()
      const inside = inflateContains(rect, x, y, 0)
      if (!inside && !inflateContains(rect, x, y, GROUP_MAGNET_PX)) continue
      const dist = inside ? 0 : distanceToRect(x, y, rect)
      if (!best || Number(inside) > Number(best.inside) || dist < best.dist) {
        best = { id, dist, inside }
      }
    }
    return best?.id || ''
  }

  function remainingIndex(groupId: string, fullIndex: number) {
    const current = locateItem(options.groups.value, dragItemId.value)
    if (!current || current.group.id !== groupId) return fullIndex
    return fullIndex > current.index ? fullIndex - 1 : fullIndex
  }

  function hitTestInsert(groupId: string, x: number, y: number) {
    const current = locateItem(options.groups.value, dragItemId.value)
    const placeholder = itemEl(dragItemId.value)
    if (
      placeholder &&
      current?.group.id === groupId &&
      inflateContains(placeholder.getBoundingClientRect(), x, y, PLACEHOLDER_KEEP_PAD_PX)
    ) {
      return current.index
    }

    const section = document.querySelector<HTMLElement>(`.home-board [data-group-id="${groupId}"]`)
    if (!section) return current?.group.id === groupId ? current.index : 0

    const group = options.groups.value.find((row) => row.id === groupId)
    const boxes: ItemBox[] = []
    section.querySelectorAll<HTMLElement>('[data-item-id]').forEach((el) => {
      const id = el.dataset.itemId
      if (!id || id === dragItemId.value) return
      const fullIndex = (group?.items || []).findIndex((item) => item.id === id)
      if (fullIndex < 0) return
      boxes.push({
        id,
        index: remainingIndex(groupId, fullIndex),
        rect: el.getBoundingClientRect(),
      })
    })

    return insertIndexAmong(boxes, x, y)
  }

  function takeItem(itemId: string) {
    for (const group of options.groups.value) {
      const items = group.items || (group.items = [])
      const index = items.findIndex((item) => item.id === itemId)
      if (index >= 0) return items.splice(index, 1)[0]
    }
    return undefined
  }

  function moveItem(itemId: string, destGroupId: string, index: number) {
    const dragged = takeItem(itemId)
    if (!dragged) return
    dragged.group_id = destGroupId
    const dest = options.groups.value.find((group) => group.id === destGroupId)
    if (!dest) return
    const items = dest.items || (dest.items = [])
    items.splice(Math.max(0, Math.min(index, items.length)), 0, dragged)
  }

  async function applyPreview(destGroupId: string, index: number) {
    if (!dragging.value || finishing) return
    const current = locateItem(options.groups.value, dragItemId.value)
    if (!current) return
    if (current.group.id === destGroupId && current.index === index) return
    const prev = captureRects()
    if (!dragging.value || finishing) return
    moveItem(dragItemId.value, destGroupId, index)
    dropGroupId.value = destGroupId
    await nextTick()
    if (!dragging.value || finishing) return
    playFlip(prev)
  }

  function applyAutoScroll() {
    const dy = edgeScrollDelta(pointer.y, window.innerHeight)
    if (!dy) return
    window.scrollBy(0, dy)
  }

  function updateGhostPosition() {
    const current = ghost.value
    if (!current || current.returning) return
    current.x = pointer.x - grabOffset.x
    current.y = pointer.y - grabOffset.y
  }

  function activateGroup(groupId: string) {
    pendingGroupId = ''
    window.clearTimeout(intentTimer)
    intentTimer = 0
    dropGroupId.value = groupId
    previewJob = { groupId, index: hitTestInsert(groupId, pointer.x, pointer.y) }
  }

  function updateDropTarget() {
    const groupId = hitTestGroup(pointer.x, pointer.y)
    hoverGroupId.value = groupId
    if (!groupId) {
      pendingGroupId = ''
      window.clearTimeout(intentTimer)
      intentTimer = 0
      return
    }

    const activeId = dropGroupId.value || originGroupId
    if (groupId === activeId) {
      pendingGroupId = ''
      window.clearTimeout(intentTimer)
      intentTimer = 0
      previewJob = { groupId, index: hitTestInsert(groupId, pointer.x, pointer.y) }
      return
    }

    if (velocity >= FAST_POINTER_PX_MS) {
      activateGroup(groupId)
      return
    }

    if (pendingGroupId === groupId) return
    pendingGroupId = groupId
    window.clearTimeout(intentTimer)
    intentTimer = window.setTimeout(() => {
      if (!dragging.value || pendingGroupId !== groupId) return
      activateGroup(groupId)
    }, GROUP_INTENT_MS)
  }

  function loop() {
    if (!dragging.value || finishing) return
    applyAutoScroll()
    updateGhostPosition()
    updateDropTarget()
    const job = previewJob
    previewJob = null
    if (job) {
      void applyPreview(job.groupId, job.index).then(() => {
        if (dragging.value && !finishing) raf = requestAnimationFrame(loop)
      })
      return
    }
    raf = requestAnimationFrame(loop)
  }

  function startDrag(item: NavItem, originEl: HTMLElement) {
    dragging.value = true
    suppressClick.value = true
    document.body.classList.add('is-item-dragging')
    options.onDragStart?.()
    const rect = originEl.getBoundingClientRect()
    ghost.value = {
      item: { ...item },
      x: rect.left,
      y: rect.top,
      width: rect.width,
      height: rect.height,
      returning: false,
    }
    dropGroupId.value = originGroupId
    hoverGroupId.value = originGroupId
    raf = requestAnimationFrame(loop)
  }

  function restoreSnapshot() {
    const prev = captureRects()
    options.groups.value = cloneGroups(snapshot)
    return prev
  }

  function wait(ms: number) {
    return new Promise<void>((resolve) => {
      window.setTimeout(resolve, prefersReducedMotion() ? 0 : ms)
    })
  }

  async function animateGhostTo(rect: DOMRect | undefined, ms: number) {
    const current = ghost.value
    if (!current) return
    if (!rect || prefersReducedMotion()) {
      ghost.value = null
      return
    }
    current.returning = true
    await nextTick()
    await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()))
    if (!ghost.value) return
    ghost.value.x = rect.left
    ghost.value.y = rect.top
    await wait(ms)
  }

  function enqueueSave(destGroupId: string, destIds: string[]) {
    saveQueue = saveQueue.then(async () => {
      try {
        await options.onCommit({ destGroupId, destIds })
      } catch {
        await options.onCommitError()
      }
    })
  }

  function releaseClickSoon() {
    window.setTimeout(() => {
      suppressClick.value = false
    }, 280)
  }

  async function finish(commit: boolean) {
    if (!dragging.value || finishing) {
      if (!dragging.value) resetSession()
      return
    }
    finishing = true
    clearTimers()
    stopLoop()
    hoverGroupId.value = ''

    const itemId = dragItemId.value
    const overGroup = Boolean(hitTestGroup(pointer.x, pointer.y))
    const changed = !sameOrder(snapshot, options.groups.value)
    const shouldCommit = commit && overGroup && changed

    if (!shouldCommit && changed) {
      const prev = restoreSnapshot()
      await nextTick()
      playFlip(prev)
      await animateGhostTo(itemEl(itemId)?.getBoundingClientRect(), RETURN_MS)
      resetSession()
      releaseClickSoon()
      return
    }

    const slot = itemEl(itemId)?.getBoundingClientRect()
    if (!shouldCommit) {
      await animateGhostTo(slot, SETTLE_MS)
      resetSession()
      releaseClickSoon()
      return
    }

    const dest = locateItem(options.groups.value, itemId)
    const destGroupId = dest?.group.id || originGroupId
    const destIds = (options.groups.value.find((group) => group.id === destGroupId)?.items || []).map(
      (item) => item.id,
    )
    await animateGhostTo(slot, SETTLE_MS)
    resetSession()
    releaseClickSoon()
    enqueueSave(destGroupId, destIds)
  }

  function onPointerMove(event: PointerEvent) {
    if (event.pointerId !== pointerId) return
    pointer = { x: event.clientX, y: event.clientY }
    const now = performance.now()
    const dt = Math.max(now - lastPointerAt, 1)
    velocity = Math.hypot(pointer.x - lastPointer.x, pointer.y - lastPointer.y) / dt
    lastPointer = pointer
    lastPointerAt = now

    if (!dragging.value) {
      const moved = Math.hypot(pointer.x - originPointer.x, pointer.y - originPointer.y)
      if (isTouchLike()) {
        if (moved > TOUCH_MOVE_CANCEL_PX) resetSession()
        return
      }
      if (moved > MOUSE_DRAG_START_PX) {
        event.preventDefault()
        beginDrag()
      }
      return
    }
    event.preventDefault()
  }

  function onPointerUp(event: PointerEvent) {
    if (event.pointerId !== pointerId) return
    pointer = { x: event.clientX, y: event.clientY }
    if (!dragging.value) {
      resetSession()
      return
    }
    event.preventDefault()
    void finish(true)
  }

  function onKeydown(event: KeyboardEvent) {
    if (event.key !== 'Escape') return
    if (!dragging.value) {
      resetSession()
      return
    }
    event.preventDefault()
    void finish(false)
  }

  function onContextMenu(event: Event) {
    if (dragging.value || longPressTimer) event.preventDefault()
  }

  function onItemPointerDown(event: PointerEvent, item: NavItem, groupId: string) {
    if (!options.enabled()) return
    if (event.button !== 0) return
    if (event.ctrlKey || event.metaKey || event.altKey) return
    if (finishing || sessionActive) return

    const el = event.currentTarget
    if (!(el instanceof HTMLElement)) return

    sessionActive = true
    pointerId = event.pointerId
    pointerKind = event.pointerType || 'mouse'
    pointer = { x: event.clientX, y: event.clientY }
    originPointer = pointer
    lastPointer = pointer
    lastPointerAt = performance.now()
    velocity = 0
    pendingItem = item
    originEl = el
    const rect = el.getBoundingClientRect()
    grabOffset = { x: event.clientX - rect.left, y: event.clientY - rect.top }
    dragItemId.value = item.id
    originGroupId = groupId
    snapshot = cloneGroups(options.groups.value)

    try {
      el.setPointerCapture(event.pointerId)
    } catch {
      /* capture is optional */
    }

    window.addEventListener('pointermove', onPointerMove, { passive: false })
    window.addEventListener('pointerup', onPointerUp)
    window.addEventListener('pointercancel', onPointerUp)
    window.addEventListener('keydown', onKeydown)
    window.addEventListener('contextmenu', onContextMenu, true)

    longPressTimer = window.setTimeout(() => {
      longPressTimer = 0
      beginDrag()
    }, LONG_PRESS_MS)
  }

  onUnmounted(() => {
    resetSession()
  })

  return {
    dragging,
    dragItemId,
    hoverGroupId,
    ghost,
    suppressClick,
    onItemPointerDown,
  }
}
