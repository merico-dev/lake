/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { useCallback, useState } from 'react'
import { ToastNotification } from '@/components/Toast'
import { DEVLAKE_ENDPOINT } from '@/utils/config'
import request from '@/utils/request'
import { Providers } from '@/data/Providers'

function useSettingsManager ({
  activeProvider,
  activeConnection,
  settings,
}) {
  const [isSaving, setIsSaving] = useState(false)
  const [isTesting, setIsTesting] = useState(false)
  const [errors, setErrors] = useState([])
  const [showError, setShowError] = useState(false)

  const buildConnectionPayload = useCallback((connection) => {
    let connectionPayload = {}
    switch (activeProvider.id) {
      case Providers.JIRA:
        connectionPayload = {
          ...connectionPayload,
          name: connection.name,
          endpoint: connection.endpoint,
          username: connection.username,
          password: connection.password,
          proxy: connection.proxy || connection.Proxy
        }
        break
      case Providers.GITHUB:
        connectionPayload = {
          ...connectionPayload,
          name: connection.name,
          endpoint: connection.endpoint,
          token: connection.token,
          proxy: connection.proxy || connection.Proxy
        }
        break
      case Providers.JENKINS:
        connectionPayload = {
          ...connectionPayload,
          name: connection.name,
          endpoint: connection.endpoint,
          username: connection.username,
          password: connection.password
        }
        break
      case Providers.GITLAB:
        connectionPayload = {
          ...connectionPayload,
          name: connection.name,
          endpoint: connection.endpoint,
          token: connection.token,
          proxy: connection.proxy || connection.Proxy
        }
        break
    }
    return connectionPayload
  }, [activeProvider.id])

  const saveSettings = useCallback(() => {
    setIsSaving(true)
    const settingsPayload = {
      ...buildConnectionPayload(activeConnection),
      ...settings,
      // DEV: true
    }

    let saveResponse = {
      success: false,
      settings: {
        ...settingsPayload
      },
      errors: []
    }

    const saveConfiguration = async (settingsPayload) => {
      try {
        setShowError(false)
        ToastNotification.clear()
        const s = await request.patch(`${DEVLAKE_ENDPOINT}/plugins/${activeProvider.id}/connections/${activeConnection.ID}`, settingsPayload)
        console.log('>> SETTINGS SAVED SUCCESSFULLY', settingsPayload, s)
        saveResponse = {
          ...saveResponse,
          success: [200, 201].includes(s.status),
          settings: { ...s.data },
          errors: s.isAxiosError ? [s.message] : []
        }
      } catch (e) {
        saveResponse.errors.push(e.message)
        setErrors(saveResponse.errors)
        console.log('>> SETTINGS FAILED TO SAVE', settingsPayload, e)
      }
    }

    saveConfiguration(settingsPayload)

    setTimeout(() => {
      if (saveResponse.success && errors.length === 0) {
        ToastNotification.show({ message: 'Instance Settings saved successfully.', intent: 'success', icon: 'small-tick' })
        setShowError(false)
        setIsSaving(false)
      } else {
        ToastNotification.show({ message: 'Instance Settings failed to save, please try again.', intent: 'danger', icon: 'error' })
        setShowError(true)
        setIsSaving(false)
      }
    }, 2000)
  }, [activeConnection, activeProvider.id, buildConnectionPayload, errors.length, settings])

  const clear = () => {

  }

  const restore = () => {

  }

  return {
    saveSettings,
    clear,
    restore,
    isSaving,
    isTesting,
    errors,
    showError,
    setIsSaving,
    setIsTesting,
    setErrors,
    setShowError,
    buildConnectionPayload,
    settings
  }
}

export default useSettingsManager
