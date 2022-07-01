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
import React, { useEffect, useState } from 'react'
import { Link, useHistory, useParams } from 'react-router-dom'
import { Icon, } from '@blueprintjs/core'
import Nav from '@/components/Nav'
import Sidebar from '@/components/Sidebar'
import AppCrumbs from '@/components/Breadcrumbs'
import Content from '@/components/Content'
import ConnectionForm from '@/pages/configure/connections/ConnectionForm'
import { integrationsData } from '@/data/integrations'
import {
  ProviderConnectionLimits,
  ProviderFormLabels,
  ProviderFormPlaceholders,
  ProviderLabels,
  Providers
} from '@/data/Providers'

import useConnectionManager from '@/hooks/useConnectionManager'
import useConnectionValidation from '@/hooks/useConnectionValidation'

import '@/styles/integration.scss'
import '@/styles/connections.scss'

export default function AddConnection () {
  const history = useHistory()
  const { providerId } = useParams()

  const [activeProvider, setActiveProvider] = useState(integrationsData.find(p => p.id === providerId))

  const {
    testConnection,
    saveConnection,
    errors,
    isSaving,
    isTesting,
    showError,
    testStatus,
    testResponse,
    name,
    endpointUrl,
    proxy,
    token,
    initialTokenStore,
    username,
    password,
    setName,
    setEndpointUrl,
    setProxy,
    setUsername,
    setPassword,
    setToken,
    setInitialTokenStore,
    fetchAllConnections,
    connectionLimitReached,
    // Providers
  } = useConnectionManager({
    activeProvider,
  })

  const {
    validate,
    errors: validationErrors,
    isValid: isValidForm
  } = useConnectionValidation({
    activeProvider,
    name,
    endpointUrl,
    proxy,
    token,
    username,
    password
  })

  const cancel = () => {
    history.push(`/integrations/${activeProvider.id}`)
  }

  // const resetForm = () => {
  //   setName(null)
  //   setEndpointUrl(null)
  //   setToken(null)
  //   setUsername(null)
  //   setPassword(null)
  // }

  useEffect(() => {
    // Selected Provider
    if (activeProvider && activeProvider.id) {
      fetchAllConnections()
      switch (activeProvider.id) {
        case Providers.JENKINS:
          setName(ProviderLabels.JENKINS)
          break
        case Providers.GITHUB:
        case Providers.GITLAB:
        case Providers.JIRA:
        default:
          setName('')
          break
      }
    }
  }, [activeProvider.id])

  useEffect(() => {
    console.log('>>>> DETECTED PROVIDER = ', providerId)
    setActiveProvider(integrationsData.find(p => p.id === providerId))
  }, [])

  return (
    <>
      <div className='container'>
        <Nav />
        <Sidebar />
        <Content>
          <main className='main'>
            <AppCrumbs
              items={[
                { href: '/', icon: false, text: 'Dashboard' },
                { href: '/integrations', icon: false, text: 'Integrations' },
                { href: `/integrations/${activeProvider.id}`, icon: false, text: `${activeProvider.name}` },
                { href: `/connections/add/${activeProvider.id}`, icon: false, text: 'Add Connection', current: true }
              ]}
            />
            <div style={{ width: '100%' }}>
              <Link style={{ float: 'right', marginLeft: '10px', color: '#777777' }} to={`/integrations/${activeProvider.id}`}>
                <Icon icon='undo' size={16} /> Go Back
              </Link>
              <div style={{ display: 'flex' }}>
                <div>
                  <span style={{ marginRight: '10px' }}>{activeProvider.icon}</span>
                </div>
                <div>
                  <h1 style={{ margin: 0 }}>
                    {activeProvider.name} Add Connection
                  </h1>
                  <p className='page-description'>Create a new connection for this provider.</p>
                </div>
              </div>
              <div className='addConnection' style={{ display: 'flex' }}>
                <ConnectionForm
                  isLocked={connectionLimitReached}
                  isValid={isValidForm}
                  validationErrors={validationErrors}
                  activeProvider={activeProvider}
                  name={name}
                  endpointUrl={endpointUrl}
                  proxy={proxy}
                  token={token}
                  initialTokenStore={initialTokenStore}
                  username={username}
                  password={password}
                  onSave={() => saveConnection({})}
                  onTest={testConnection}
                  onCancel={cancel}
                  onValidate={validate}
                  onNameChange={setName}
                  onEndpointChange={setEndpointUrl}
                  onProxyChange={setProxy}
                  onTokenChange={setToken}
                  onUsernameChange={setUsername}
                  onPasswordChange={setPassword}
                  isSaving={isSaving}
                  isTesting={isTesting}
                  testStatus={testStatus}
                  testResponse={testResponse}
                  errors={errors}
                  showError={showError}
                  authType={[Providers.JENKINS, Providers.JIRA].includes(activeProvider.id) ? 'plain' : 'token'}
                  sourceLimits={ProviderConnectionLimits}
                  labels={ProviderFormLabels[activeProvider.id]}
                  placeholders={ProviderFormPlaceholders[activeProvider.id]}
                />
              </div>
              {/* {validationErrors.length > 0 && (
                <FormValidationErrors errors={validationErrors} />
              )} */}
            </div>
          </main>
        </Content>
      </div>
    </>
  )
}
