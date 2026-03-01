<template>
  <div class="mx-auto max-w-[1400px] px-5 py-10 md:px-10">
    <header class="mb-10 rise flex items-end justify-between">
      <div>
        <div class="flex items-center gap-3">
          <div class="h-2 w-2 rounded-full bg-accent" style="animation: pulse-dot 2s ease-in-out infinite;"></div>
          <p class="mono text-2xs uppercase tracking-[0.3em] text-accent/70">live telemetry</p>
        </div>
        <h1 class="mt-3 text-4xl font-bold tracking-tight md:text-5xl">diskmon</h1>
      </div>
      <p class="mono hidden text-2xs text-[var(--text-tertiary)] md:block">disk health monitoring</p>
    </header>
    <router-view />
  </div>
</template>
