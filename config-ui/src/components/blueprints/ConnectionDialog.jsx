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
import React, { useEffect, useState, useRef } from 'react'
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
  ProviderFormPlaceholders,
  ProviderIcons,
} from '@/data/Providers'
import InputValidationError from '@/components/validation/InputValidationError'
import ContentLoader from '@/components/loaders/ContentLoader'
import { NullBlueprintConnection } from '@/data/NullBlueprintConnection'

const Modes = {
  CREATE: 'create',
  EDIT: 'edit',
}

const ConnectionDialog = (props) => {
  const {
    isOpen = false,
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
    errors = [],
  } = props

  const connectionNameRef = useRef()
  const connectionEndpointRef = useRef()

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

  useEffect(() => {}, [datasource])

  useEffect(() => {
    if (connection?.id) {
      setMode(Modes.EDIT)
      setDatasource(
        dataSourcesList.find((d) => d.value === connection.provider)
      )
    } else {
      setMode(Modes.CREATE)
    }
  }, [connection])

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
              title='Loading Connection...'
              elevation={Elevation.ZERO}
              message='Please wait for connection.'
            />
          ) : (
            <>
              <div className='manage-connection'>
                <div className='formContainer'>
                  <FormGroup
                    disabled={isTesting || isSaving || isLocked}
                    readOnly={[
                      Providers.GITHUB,
                      Providers.GITLAB,
                      Providers.JENKINS,
                    ].includes(connection.id)}
                    label=''
                    inline={true}
                    labelFor='connection-name'
                    className='formGroup-inline'
                    contentClassName='formGroupContent'
                  >
                    <Label style={{ display: 'inline', marginRight: 0 }}>
                      {labels ? labels.name : <>Connection&nbsp;Name</>}
                      <span className='requiredStar'>*</span>
                    </Label>
                    <p>
                      Give your connection a unique name to help you identify it
                      in the future.
                    </p>
                    <InputGroup
                      id='connection-name'
                      inputRef={connectionNameRef}
                      disabled={isTesting || isSaving || isLocked}
                      readOnly={[
                        Providers.GITHUB,
                        Providers.GITLAB,
                        Providers.JENKINS,
                      ].includes(connection.provider)}
                      placeholder={
                        placeholders ? placeholders.name : 'Enter Instance Name'
                      }
                      value={connection.title}
                      onChange={(e) => onNameChange(e.target.value)}
                      className={`connection-name-input ${
                        stateErrored === 'connection-name'
                          ? 'invalid-field'
                          : ''
                      }`}
                      leftIcon={
                        [
                          Providers.GITHUB,
                          Providers.GITLAB,
                          Providers.JENKINS,
                        ].includes(connection.provider)
                          ? 'lock'
                          : null
                      }
                      inline={true}
                      rightElement={
                        <InputValidationError
                          error={getFieldError('Connection')}
                          elementRef={connectionNameRef}
                          onError={activateErrorStates}
                          onSuccess={() => setStateErrored(null)}
                          validateOnFocus
                        />
                      }
                      // fill
                    />
                  </FormGroup>
                </div>

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
                      readOnly={connection?.id && mode === Modes.EDIT}
                    >
                      <Button
                        disabled={connection?.id && mode === Modes.EDIT}
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

                {[Providers.JIRA].includes(datasource.value) && (
                  <div className='formContainer'>
                    <FormGroup
                      disabled={isTesting || isSaving || isLocked}
                      label=''
                      inline={true}
                      labelFor='connection-endpoint'
                      className='formGroup-inline formGroup-full'
                      contentClassName='formGroupContent'
                    >
                      <Label>
                        {labels ? labels.endpoint : <>Endpoint&nbsp;URL</>}
                        <span className='requiredStar'>*</span>
                      </Label>
                      <InputGroup
                        id='connection-endpoint'
                        inputRef={connectionEndpointRef}
                        disabled={isTesting || isSaving || isLocked}
                        placeholder={
                          placeholders
                            ? placeholders.endpoint
                            : 'Enter Endpoint URL'
                        }
                        value={endpointUrl}
                        onChange={(e) => onEndpointChange(e.target.value)}
                        className={`endpoint-url-input ${
                          stateErrored === 'connection-endpoint'
                            ? 'invalid-field'
                            : ''
                        }`}
                        fill
                        rightElement={
                          <InputValidationError
                            error={getFieldError('Endpoint')}
                            elementRef={connectionEndpointRef}
                            onError={activateErrorStates}
                            onSuccess={() => setStateErrored(null)}
                            validateOnFocus
                          />
                        }
                      />
                    </FormGroup>
                  </div>
                )}
              </div>
            </>
          )}
        </div>
        <div className={Classes.DIALOG_FOOTER}>
          <div className={Classes.DIALOG_FOOTER_ACTIONS}>
            <Button
              disabled={isSaving || isTesting}
              onClick={() => onTest(false)}
            >
              Test Connection
            </Button>
            <Button
              disabled={isSaving || !isValid || isTesting}
              icon='cloud-upload'
              intent={Intent.PRIMARY}
              onClick={() => onSave(connection ? connection.id : null)}
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
