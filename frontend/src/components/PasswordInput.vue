<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'

defineOptions({ inheritAttrs: false })

defineProps<{
  autocomplete?: string
  placeholder?: string
  disabled?: boolean
  inputClass?: string
}>()

const model = defineModel<string>({ default: '' })
const { t } = useI18n()
const visible = ref(false)

function toggle() {
  visible.value = !visible.value
}
</script>

<template>
  <div class="password-wrap">
    <input
      v-model="model"
      :class="inputClass"
      :type="visible ? 'text' : 'password'"
      :autocomplete="autocomplete"
      :placeholder="placeholder"
      :disabled="disabled"
      spellcheck="false"
      v-bind="$attrs"
    />
    <button
      type="button"
      class="toggle"
      :disabled="disabled"
      :aria-label="visible ? t('common.hidePassword') : t('common.showPassword')"
      :title="visible ? t('common.hidePassword') : t('common.showPassword')"
      @click="toggle"
    >
      <Icon :icon="visible ? 'mdi:eye-off-outline' : 'mdi:eye-outline'" width="18" />
    </button>
  </div>
</template>

<style scoped>
.password-wrap {
  position: relative;
  display: block;
  width: 100%;
}

.password-wrap input {
  width: 100%;
  box-sizing: border-box;
  padding-right: 40px;
}

.toggle {
  position: absolute;
  right: 4px;
  top: 50%;
  transform: translateY(-50%);
  width: 32px;
  height: 32px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: inherit;
  opacity: 0.62;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.toggle:hover:not(:disabled) {
  opacity: 1;
  background: color-mix(in srgb, currentColor 10%, transparent);
}

.toggle:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
</style>
