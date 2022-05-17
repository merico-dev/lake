import { useState, useEffect, useCallback } from 'react'
import {
  useHistory
} from 'react-router-dom'
import { ToastNotification } from '@/components/Toast'
import { DEVLAKE_ENDPOINT } from '@/utils/config'
import request from '@/utils/request'
import { Providers } from '@/data/Providers'

function useSettingsManager ({
  activeProvider,
  activeConnection,
  settings,
}) {
  const history = useHistory()

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
          endPoing: connection.endpoint,
          basicAuthEncoded: connection.basicAuthEncoded,
          proxy: connection.proxy || connection.Proxy
        }
        break
      case Providers.GITHUB:
        connectionPayload = {
          ...connectionPayload,
          name: connection.name,
          endpoint: connection.endpoint,
          auth: connection.auth,
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
          auth: connection.auth,
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
