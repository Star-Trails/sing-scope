<template>
  <div
    class="bg-base-200 home-page flex size-full overflow-hidden"
    :class="sidebarLayoutCollapsed ? 'sidebar-collapsed' : 'sidebar-expanded'"
  >
    <div
      v-if="!isMiddleScreen"
      class="relative z-40 flex-none h-full transition-all duration-300"
      :class="sidebarLayoutCollapsed ? 'w-16' : 'w-64'"
    >
      <SideBar
        class="h-full"
        @transitionend="syncSidebarLayoutState"
      />
    </div>
    <RouterView v-slot="{ Component, route }">
      <div
        class="relative flex-1 h-full overflow-hidden"
        ref="swiperRef"
      >
        <div class="flex h-full w-full flex-col overflow-y-auto">
          <Transition
            :name="(route.meta.transition as string) || 'fade'"
            v-if="isMiddleScreen"
          >
            <Component :is="Component" />
          </Transition>
          <Transition
            v-else
            name="page"
            mode="out-in"
          >
            <Component :is="Component" />
          </Transition>
        </div>

        <template v-if="isMiddleScreen">
          <div
            class="bg-base-100/20 dock dock-xs z-10 h-14 w-auto"
            :style="{
              padding: '0',
              bottom: 'calc(var(--spacing) * 2 + env(safe-area-inset-bottom))',
            }"
            ref="dockRef"
          >
            <button
              v-for="r in renderRoutes"
              :key="r"
              @click="router.push({ name: r, replace: true })"
              class="h-14 flex-col items-center justify-center pt-2"
              :class="r === route.name && 'dock-active'"
            >
              <component
                :is="ROUTE_ICON_MAP[r]"
                class="h-5 w-5 flex-shrink-0"
              />
              <span class="dock-label">
                {{ $t(r) }}
              </span>
            </button>
          </div>
        </template>
      </div>
    </RouterView>

    <ConfirmDialogHost />
    <BackendSwitchToast />
  </div>
</template>

<script setup lang="ts">
import BackendSwitchToast from '@/components/common/BackendSwitchToast.vue'
import ConfirmDialogHost from '@/components/common/ConfirmDialogHost.vue'
import SideBar from '@/components/sidebar/SideBar.vue'
import { useSwipeRouter } from '@/composables/swipe'
import { ROUTE_ICON_MAP } from '@/constant'
import { renderRoutes } from '@/helper'
import { isMiddleScreen } from '@/helper/utils'
import router from '@/router'
import { isSidebarCollapsed } from '@/store/settings'
import { nextTick, ref, watch } from 'vue'

const sidebarLayoutCollapsed = ref(isSidebarCollapsed.value)

const syncSidebarLayoutState = () => {
  sidebarLayoutCollapsed.value = isSidebarCollapsed.value
}

watch(isSidebarCollapsed, (collapsed) => {
  if (!collapsed) {
    sidebarLayoutCollapsed.value = false
    return
  }
  nextTick(syncSidebarLayoutState)
})

const { swiperRef } = useSwipeRouter()
</script>

<style scoped>
.page-enter-active,
.page-leave-active {
  transition: opacity 0.15s ease-out;
}

.page-enter-from,
.page-leave-to {
  opacity: 0;
}
</style>
