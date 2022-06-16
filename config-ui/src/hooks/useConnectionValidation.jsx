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
import { useState, useEffect, useCallback } from 'react'
import {
  Providers,
} from '@/data/Providers'

function useConnectionValidation ({
  activeProvider,
  name,
  endpointUrl,
  proxy,
  token,
  username,
  password
}) {
  const [errors, setErrors] = useState([])
  const [isValid, setIsValid] = useState(false)
  const [validURIs, setValidURIs] = useState([
    'http://',
    'https://'
  ])

  const clear = () => {
    setErrors([])
  }

  const validate = useCallback(() => {
    const errs = []
    console.log('>> VALIDATING PROVIDER ID: ', activeProvider?.id)
    console.log('>> RUNNING FORM VALIDATIONS AGAINST FIELD VALUES...')
    console.log(
      'NAME', name,
      'ENDPOINT URL', endpointUrl,
      'PROXY URL', proxy,
      'TOKEN', token,
      'USERNAME', username,
      'PASSWORD', password
    )

    if (!name) {
      errs.push('Connection name is required')
    }

    if (name && name.length <= 2) {
      errs.push('Connection name too short/incomplete')
    }

    if (!endpointUrl || endpointUrl.length <= 2) {
      errs.push('Endpoint URL is required')
    }

    if (!endpointUrl?.startsWith('http')) {
      errs.push('Endpoint URL must be valid HTTP/S protocol')
    }

    if (!endpointUrl?.endsWith('/')) {
      errs.push('Endpoint URL must end in trailing slash (/)')
    }

    if (proxy && proxy !== '' && !validURIs.some(uri => proxy?.startsWith(uri))) {
      errs.push('Proxy URL must be valid HTTP/S protocol')
    }

    switch (activeProvider?.id) {
      case Providers.GITHUB:
      case Providers.GITLAB:
        if (!token || token.length <= 2) {
          errs.push('Authentication token(s) are required')
        }
        break
      case Providers.JIRA:
      case Providers.JENKINS:
        if (!username || username.length <= 2) {
          errs.push('Username is required')
        }
        if (!password || password.length <= 2) {
          errs.push('Password is required')
        }
        break
    }

    setErrors(errs)
  }, [
    name,
    endpointUrl,
    proxy,
    token,
    username,
    password,
    activeProvider,
    validURIs
  ])

  useEffect(() => {
    console.log('>>> CONNECTION FORM ERRORS...', errors)
    setIsValid(errors.length === 0)
    if (errors.length > 0) {
      // ToastNotification.clear()
      // ToastNotification.show({ message: errors[0], intent: 'danger', icon: 'warning-sign' })
    }
  }, [errors])

  return {
    errors,
    isValid,
    validate,
    clear,
    setValidURIs
  }
}

export default useConnectionValidation
