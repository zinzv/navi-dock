<script setup lang="ts">
</script>

<template>
  <div class="scene" aria-hidden="true">
    <svg viewBox="0 0 640 720" fill="none" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id="desk" x1="80" y1="120" x2="560" y2="640" gradientUnits="userSpaceOnUse">
          <stop stop-color="var(--scene-stroke)" stop-opacity="1" />
          <stop offset="1" stop-color="var(--scene-stroke-soft)" stop-opacity="0.75" />
        </linearGradient>
        <radialGradient
          id="cone"
          cx="0"
          cy="0"
          r="1"
          gradientUnits="userSpaceOnUse"
          gradientTransform="translate(430 170) rotate(90) scale(260 120)"
        >
          <stop stop-color="var(--scene-glow)" stop-opacity="0.55" />
          <stop offset="1" stop-color="var(--scene-glow)" stop-opacity="0" />
        </radialGradient>
      </defs>
      <g transform="translate(20 31) scale(.94)">
        <path class="scene-base scene-base-outer" d="M120 520 L320 420 L560 520 L360 620 Z" stroke="url(#desk)" stroke-width="2" />
        <path class="scene-base scene-base-inner" d="M180 500 L330 430 L480 500 L330 570 Z" stroke="url(#desk)" stroke-width="1.6" />
      </g>
      <rect x="250" y="360" width="180" height="110" rx="8" stroke="url(#desk)" stroke-width="2" opacity="0.95" />
      <path d="M250 380 H430" stroke="var(--scene-stroke)" stroke-width="1.4" opacity="0.7" />
      <path d="M270 400 H410 M270 415 H390 M270 430 H400" stroke="var(--scene-stroke)" stroke-width="1.4" opacity="0.65" />
      <g class="scene-top" transform="translate(0 12)">
        <path d="M430 170 L470 250 L390 250 Z" fill="url(#cone)" />
        <circle class="lamp-glow" cx="430" cy="160" r="12" fill="var(--scene-glow)" opacity="1" />
        <path d="M430 170 V250" stroke="var(--scene-stroke)" stroke-width="2" opacity="0.9" />
        <path d="M360 250 H500" stroke="var(--scene-stroke)" stroke-width="1.6" opacity="0.7" />
        <circle cx="520" cy="220" r="28" stroke="var(--scene-stroke)" stroke-width="1.6" opacity="0.75" />
        <circle cx="520" cy="220" r="8" fill="var(--scene-stroke)" opacity="0.55" />
        <path d="M140 280 C180 240, 220 300, 260 270" stroke="var(--scene-stroke)" stroke-width="1.5" opacity="0.6" />
        <circle cx="160" cy="260" r="5" fill="var(--scene-stroke)" opacity="0.75" />
        <circle cx="210" cy="285" r="4" fill="var(--scene-stroke)" opacity="0.7" />
        <circle cx="250" cy="268" r="5" fill="var(--scene-stroke)" opacity="0.75" />
      </g>
      <g transform="translate(-8 10)">
        <rect x="470" y="300" width="90" height="70" rx="8" stroke="var(--scene-stroke)" stroke-width="1.8" opacity="0.9" />
        <path d="M485 320 H545 M485 335 H530 M485 350 H540" stroke="var(--scene-stroke)" stroke-width="1.4" opacity="0.65" />
        <text x="470" y="290" fill="var(--scene-stroke)" opacity="0.75" font-size="11" font-family="Segoe UI, sans-serif" font-weight="600">NETWORKS</text>
      </g>
      <text x="250" y="350" fill="var(--scene-stroke)" opacity="0.7" font-size="11" font-family="Segoe UI, sans-serif" font-weight="600">CPU</text>
    </svg>
  </div>
</template>
