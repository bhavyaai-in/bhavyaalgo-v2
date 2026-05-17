<script setup>
import { useUiStore } from '../stores/ui.js'

const ui = useUiStore()
</script>

<template>
  <div v-if="ui.isLoading" id="preloader">
    <div class="loader-bg" />
    <div class="loader-ring">
      <svg viewBox="0 0 72 72">
        <circle class="ring-track" cx="36" cy="36" r="32" />
        <circle class="ring-arc a2" cx="36" cy="36" r="32" />
        <circle class="ring-arc a1" cx="36" cy="36" r="32" />
      </svg>
      <div class="ring-center-dot" />
    </div>
  </div>
</template>

<style scoped>
#preloader {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
}

.loader-bg {
  position: absolute;
  inset: 0;
  background: hsl(var(--background) / .4);
}

.loader-ring {
  position: relative;
  width: 72px;
  height: 72px;
}

.loader-ring svg {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  animation: ring-spin 2s linear infinite;
}

.ring-track {
  fill: none;
  stroke: hsl(var(--border));
  stroke-width: 2;
}

.ring-arc {
  fill: none;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-dasharray: 180 226;
  stroke-dashoffset: 0;
  animation: arc-chase 2s ease-in-out infinite;
}

.ring-arc.a1 {
  stroke: hsl(var(--primary));
  filter: drop-shadow(0 0 1px hsl(var(--primary)));
}

.ring-arc.a2 {
  stroke: hsl(var(--primary) / .3);
  stroke-dasharray: 60 346;
  animation: arc-chase2 2s ease-in-out infinite;
}

.ring-center-dot {
  position: absolute;
  top: 50%; left: 50%;
  transform: translate(-50%,-50%);
  width: 6px; height: 6px;
  background: hsl(var(--primary));
  border-radius: 50%;
  /* box-shadow: 0 0 14px hsl(var(--primary)), 0 0 28px hsl(var(--primary) / .3); */
  animation: dot-breathe 2s ease-in-out infinite;
}

@keyframes ring-spin { to { transform: rotate(360deg); } }

@keyframes arc-chase {
  0%   { stroke-dashoffset: 0; }
  50%  { stroke-dashoffset: -180; }
  100% { stroke-dashoffset: -406; }
}

@keyframes arc-chase2 {
  0%   { stroke-dashoffset: 0; }
  50%  { stroke-dashoffset: -80; }
  100% { stroke-dashoffset: -406; }
}

@keyframes dot-breathe {
  0%,100% { transform: translate(-50%,-50%) scale(1); opacity: .7; }
  50%      { transform: translate(-50%,-50%) scale(1.5); opacity: 1; }
}
</style>
