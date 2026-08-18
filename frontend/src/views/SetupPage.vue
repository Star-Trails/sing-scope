<template>
  <div
    class="bg-base-200/50 h-full w-full items-center justify-center overflow-auto sm:flex select-none"
    @keydown.enter="handleSubmit(form)"
  >
    <div class="absolute top-4 right-4 max-sm:hidden">
      <DashboardSettings />
    </div>
    <div class="absolute right-4 bottom-4 max-sm:hidden">
      <LanguageSelect />
    </div>
    <div
      class="border-base-200 bg-base-100 mx-auto flex w-96 max-w-[90%] flex-col gap-3 rounded-2xl border px-6 py-5 shadow-lg max-sm:my-4"
    >
      <div class="flex items-center space-x-2 mb-1">
        <div class="w-6 h-6 rounded-lg bg-primary/20 flex items-center justify-center text-primary font-black text-xs">
          S
        </div>
        <h1 class="text-base font-bold text-base-content">sing-box API Connection</h1>
      </div>

      <div class="flex gap-2">
        <div class="flex w-24 flex-none flex-col gap-1">
          <label class="text-xs font-semibold text-base-content/70">{{ $t('protocol') }}</label>
          <select
            class="select select-sm select-bordered w-full text-xs font-mono"
            v-model="form.protocol"
          >
            <option value="http">HTTP</option>
            <option value="https">HTTPS</option>
          </select>
        </div>
        <div class="flex min-w-0 flex-1 flex-col gap-1">
          <label class="text-xs font-semibold text-base-content/70">{{ $t('host') }}</label>
          <TextInput
            class="w-full font-mono text-xs"
            name="username"
            autocomplete="username"
            v-model="form.host"
            placeholder="127.0.0.1"
          />
        </div>
        <div class="flex w-20 flex-none flex-col gap-1">
          <label class="text-xs font-semibold text-base-content/70">{{ $t('port') }}</label>
          <TextInput
            class="w-full font-mono text-xs"
            v-model="form.port"
            placeholder="9090"
          />
        </div>
      </div>

      <div class="flex min-w-0 flex-1 flex-col gap-1">
        <label class="truncate text-xs font-semibold text-base-content/70">{{ $t('label') }} ({{ $t('optional') }})</label>
        <TextInput
          class="w-full text-xs"
          v-model="form.label"
          placeholder="Local sing-box"
        />
      </div>

      <div class="flex flex-col gap-1">
        <label class="text-xs font-semibold text-base-content/70">API Secret / Bearer Token</label>
        <input
          type="password"
          class="input input-sm input-bordered w-full font-mono text-xs"
          v-model="form.password"
          placeholder="Bearer Secret (leave blank if disabled)"
        />
      </div>

      <button
        class="btn btn-primary btn-sm w-full mt-2 font-bold shadow-md"
        @click="handleSubmit(form)"
      >
        {{ $t('submit') }}
      </button>

      <template v-if="backendList.length">
        <div class="text-base-content/50 mt-2 text-xs font-semibold uppercase tracking-wider">{{ $t('backend') }}</div>
        <Draggable
          class="-mr-2 flex max-h-48 flex-1 flex-col gap-1 overflow-y-auto pr-2"
          v-model="backendList"
          group="list"
          handle=".drag-handle"
          :animation="150"
          :item-key="'uuid'"
        >
          <template #item="{ element }">
            <div
              :key="element.uuid"
              class="group hover:bg-base-200 flex items-center gap-1 rounded-lg pr-1 transition-colors p-1"
            >
              <ChevronUpDownIcon
                class="drag-handle text-base-content/30 ml-1 h-4 w-4 flex-none cursor-grab"
              />
              <button
                class="min-w-0 flex-1 truncate py-1 text-left text-xs font-mono font-medium"
                @click="selectBackend(element.uuid)"
              >
                {{ getLabelFromBackend(element) }}
              </button>
              <button
                class="btn btn-circle btn-ghost btn-xs text-base-content/40 hover:text-base-content opacity-0 group-hover:opacity-100"
                @click="editBackend(element)"
              >
                <PencilIcon class="h-3.5 w-3.5" />
              </button>
              <button
                class="btn btn-circle btn-ghost btn-xs text-base-content/40 hover:text-error opacity-0 group-hover:opacity-100"
                @click="removeBackend(element.uuid)"
              >
                <TrashIcon class="h-3.5 w-3.5" />
              </button>
            </div>
          </template>
        </Draggable>
      </template>

      <div class="mt-4 sm:hidden">
        <LanguageSelect />
      </div>
    </div>

    <EditBackendModal
      v-model="showEditModal"
      :default-backend-uuid="editingBackendUuid"
    />
  </div>
</template>

<script setup lang="ts">
import { isSingboxChannelAvailable } from '@/assembly/backend'
import DashboardSettings from '@/components/common/DashboardSettings.vue'
import TextInput from '@/components/common/TextInput.vue'
import EditBackendModal from '@/components/settings/backend/EditBackendModal.vue'
import LanguageSelect from '@/components/settings/general/LanguageSelect.vue'
import { ROUTE_NAME } from '@/constant'
import { showNotification } from '@/helper/notification'
import { getBackendFromUrl, getLabelFromBackend } from '@/helper/utils'
import router from '@/router'
import { activeUuid, addBackend, backendList, removeBackend } from '@/store/setup'
import type { Backend, BackendType } from '@/types'
import {
  ChevronUpDownIcon,
  PencilIcon,
  TrashIcon,
} from '@heroicons/vue/24/outline'
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Draggable from 'vuedraggable'

const { t } = useI18n()

const form = reactive({
  type: 'singbox' as BackendType,
  protocol: 'http',
  host: '127.0.0.1',
  port: '9090',
  secondaryPath: '',
  password: '',
  label: '',
})

const showEditModal = ref(false)
const editingBackendUuid = ref('')
const isManualSetupRoute = () => router.currentRoute.value.query.setupMode === 'manual'
const isEditBackendRoute = () => typeof router.currentRoute.value.query.editBackend === 'string'

watch(
  () => router.currentRoute.value.query.editBackend,
  (backendUuid) => {
    if (backendUuid && typeof backendUuid === 'string') {
      editingBackendUuid.value = backendUuid
      showEditModal.value = true
      router.replace({ query: {} })
    }
  },
  { immediate: true },
)

const selectBackend = (uuid: string) => {
  activeUuid.value = uuid
  router.push({ name: ROUTE_NAME.overview })
}

const editBackend = (backend: Backend) => {
  editingBackendUuid.value = backend.uuid
  showEditModal.value = true
}

type SetupForm = Omit<Backend, 'uuid'>

const finishLogin = () => {
  router.push({ name: ROUTE_NAME.overview })
}

const handleSubmit = async (setupForm: SetupForm, quiet = false) => {
  const { protocol, host, port } = setupForm
  if (!protocol || !host || !port) {
    if (!quiet) alert('Please fill in all the fields.')
    return
  }

  setupForm.type = 'singbox'
  addBackend(setupForm)

  const appService = (window as any).go?.app?.AppService
  if (appService?.ConnectServer) {
    const url = `${setupForm.protocol || 'http'}://${setupForm.host}:${setupForm.port || '9090'}`
    appService.ConnectServer(url, setupForm.password || '')
  }

  finishLogin()
}

const backend = isManualSetupRoute() || isEditBackendRoute() ? null : getBackendFromUrl()

if (backend) {
  handleSubmit(backend)
} else if (backendList.value.length === 0) {
  handleSubmit(form, true)
}
</script>
