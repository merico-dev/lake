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
import React, { useEffect, useState, useRef, useCallback } from 'react'
import dayjs from '@/utils/time'
import {
  Button,
  ButtonGroup,
  Classes,
  Colors,
  Dialog,
  Elevation,
  FormGroup,
  Icon,
  InputGroup,
  Intent,
  Label,
  MenuItem,
  Popover,
  Position,
  Radio,
  RadioGroup,
  Switch,
  Tooltip,
} from '@blueprintjs/core'
import { Select } from '@blueprintjs/select'
import {
  Providers,
  ProviderTypes,
  ProviderLabels,
  ProviderFormLabels,
  ProviderFormPlaceholders,
  ProviderConnectionLimits,
  ProviderIcons,
} from '@/data/Providers'
import { NullBlueprintConnection } from '@/data/NullBlueprintConnection'
import InputValidationError from '@/components/validation/InputValidationError'
import ContentLoader from '@/components/loaders/ContentLoader'
import ConnectionForm from '@/pages/configure/connections/ConnectionForm'

const Modes = {
  CREATE: 'create',
  EDIT: 'edit',
}

const ConnectionDialog = (props) => {
  const {
    isOpen = false,
    activeProvider,
    integrations = [],
    setProvider = () => {},
    connection = NullBlueprintConnection,
    name,
    endpointUrl,
    proxy,
    token,
    username,
    password,
    isLocked = false,
    isLoading = false,
    isTesting = false,
    isSaving = false,
    isValid = false,
    // editMode = false,
    dataSourcesList = [
      {
        id: 1,
        name: Providers.JIRA,
        title: ProviderLabels[Providers.JIRA.toUpperCase()],
        value: Providers.JIRA,
      },
      {
        id: 2,
        name: Providers.GITHUB,
        title: ProviderLabels[Providers.GITHUB.toUpperCase()],
        value: Providers.GITHUB,
      },
      {
        id: 3,
        name: Providers.GITLAB,
        title: ProviderLabels[Providers.GITLAB.toUpperCase()],
        value: Providers.GITLAB,
      },
      {
        id: 4,
        name: Providers.JENKINS,
        title: ProviderLabels[Providers.JENKINS.toUpperCase()],
        value: Providers.JENKINS,
      },
    ],
    labels = ProviderLabels[connection.provider],
    placeholders = ProviderFormPlaceholders[connection.provider],
    onTest = () => {},
    onSave = () => {},
    onClose = () => {},
    onCancel = () => {},
    onValidate = () => {},
    onNameChange=() => {},
    onEndpointChange=() => {},
    onProxyChange=() => {},
    onTokenChange=() => {},
    onUsernameChange=() => {},
    onPasswordChange=() => {},
    showConnectionError = false,
    testStatus,
    errors = [],
    validationErrors = [],
    // authType,
    // showLimitWarning = false
  } = props

  {/* const connectionNameRef = useRef()
  const connectionEndpointRef = useRef() */}

  const [datasource, setDatasource] = useState(
    connection?.id
      ? dataSourcesList.find((d) => d.value === connection.provider)
      : dataSourcesList[0]
  )

  const [stateErrored, setStateErrored] = useState(false)

  const [mode, setMode] = useState(Modes.CREATE)

  const getFieldError = (fieldId) => {
    return errors.find((e) => e.includes(fieldId))
  }

  const activateErrorStates = (elementId) => {
    setStateErrored(elementId || false)
  }
  
  const getConnectionStatusIcon = useCallback(() => {
    let i = <Icon icon='full-circle' size='10' color={Colors.RED3} />
    switch (testStatus) {
      case 1:
        i = <Icon icon='full-circle' size='10' color={Colors.GREEN3} />
        break
      case 2:
        i = <Icon icon='full-circle' size='10' color={Colors.RED3} />
        break
      case 0:
      default:
        i = <Icon icon='full-circle' size='10' color={Colors.GRAY3} />
        break
    }
    return i
  }, [testStatus])

  useEffect(() => {
    if (connection?.id !== null && connection?.id !== undefined) {
      setMode(Modes.EDIT)
      setDatasource(
        dataSourcesList.find((d) => d.value === connection.provider)
      )
    } else {
      setMode(Modes.CREATE)
    }
  }, [connection])
  
  useEffect(() => {
    console.log('>>> DATASOURCE CHANGED....', datasource)
    setProvider(integrations.find(p => p.id === datasource.value))
  }, [datasource])

  useEffect(() => {
    
  }, [testStatus])

  return (
    <>
      <Dialog
        className='dialog-manage-connection'
        icon={mode === Modes.EDIT ? 'edit' : 'add'}
        title={
          mode === Modes.EDIT
            ? `Modify ${connection?.name} [#${connection?.value}]`
            : 'Create a New Data Connection'
        }
        isOpen={isOpen}
        onClose={onClose}
        onClosed={() => {}}
        style={{ backgroundColor: '#ffffff' }}
      >
        <div className={Classes.DIALOG_BODY}>
          {isLoading || isSaving ? (
            <ContentLoader
              title={`${isSaving ? 'Saving' : 'Loading'} Connection...`}
              elevation={Elevation.ZERO}
              message='Please wait.'
            />
          ) : (
            <>
              <div className='manage-connection'>

                <div className='formContainer'>
                  <FormGroup
                    disabled={isTesting || isSaving || isLocked}
                    label=''
                    inline={true}
                    labelFor='selector-datasource'
                    className='formGroup-inline'
                    contentClassName='formGroupContent'
                  >
                    <Label style={{ display: 'inline', marginRight: 0 }}>
                      {labels ? labels.datasource : <>Data Source</>}
                      <span className='requiredStar'>*</span>
                    </Label>
                    <Select
                      popoverProps={{ usePortal: false }}
                      className='selector-datasource'
                      id='selector-datasource'
                      inline={false}
                      fill={true}
                      items={dataSourcesList}
                      activeItem={datasource}
                      itemPredicate={(query, item) =>
                        item.title.toLowerCase().indexOf(query.toLowerCase()) >=
                        0
                      }
                      itemRenderer={(item, { handleClick, modifiers }) => (
                        <MenuItem
                          active={modifiers.active}
                          key={item.value}
                          label={item.value}
                          onClick={handleClick}
                          text={item.title}
                        />
                      )}
                      noResults={
                        <MenuItem disabled={true} text='No data sources.' />
                      }
                      onItemSelect={(item) => {
                        setDatasource(item)
                      }}
                      readOnly={connection?.id !== null && mode === Modes.EDIT}
                    >
                      <Button
                        disabled={connection?.id !== null && mode === Modes.EDIT}
                        className='btn-select-datasource'
                        intent={Intent.NONE}
                        style={{ maxWidth: '260px' }}
                        text={
                          datasource
                            ? `${datasource.title}`
                            : '< Select Datasource >'
                        }
                        rightIcon='double-caret-vertical'
                        fill
                        style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                        }}
                      />
                    </Select>
                  </FormGroup>
                </div>

                
                <div className='connection-form-wrapper' style={{ display: 'flex' }}>
                 <ConnectionForm
                   isValid={isValid}
                   validationErrors={validationErrors}
                   activeProvider={activeProvider}
                   name={name}
                   endpointUrl={endpointUrl}
                   proxy={proxy}
                   token={token}
                   username={username}
                   password={password}
                   onSave={onSave}
                   onTest={onTest}
                   onCancel={onCancel}
                   onValidate={onValidate}
                   onNameChange={onNameChange}
                   onEndpointChange={onEndpointChange}
                   onProxyChange={onProxyChange}
                   onTokenChange={onTokenChange}
                   onUsernameChange={onUsernameChange}
                   onPasswordChange={onPasswordChange}
                   isSaving={isSaving}
                   isTesting={isTesting}
                   testStatus={testStatus}
                   errors={errors}
                   showError={showConnectionError}
                   authType={[Providers.JENKINS, Providers.JIRA].includes(activeProvider?.id) ? 'plain' : 'token'}
                   showLimitWarning={false}
                   sourceLimits={ProviderConnectionLimits}
                   labels={ProviderFormLabels[activeProvider?.id]}
                   placeholders={ProviderFormPlaceholders[activeProvider?.id]}
                   enableActions={false}
                   // formGroupClassName='formGroup-inline'
                   showHeadline={false}
                 />
                </div>
                
              </div>
            </>
          )}
        </div>
        <div className={Classes.DIALOG_FOOTER}>
          <div className={Classes.DIALOG_FOOTER_ACTIONS}>
            <Button
              className='btn-test'
              icon={getConnectionStatusIcon()}
              disabled={isSaving || isTesting}
              onClick={() => onTest(false)}
              loading={isTesting}
            >
              Test Connection
            </Button>
            <Button
              className='btn-save'
              disabled={isSaving || !isValid || isTesting}
              icon='cloud-upload'
              intent={Intent.PRIMARY}
              onClick={() => onSave(connection ? connection.id : null)}
              loading={isSaving}
            >
              {'Save Connection'}
            </Button>
          </div>
        </div>
      </Dialog>
    </>
  )
}

export default ConnectionDialog
