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
import React, { Fragment, useEffect, useState, useCallback } from 'react'
import {
  Button,
  Icon,
  Intent,
  InputGroup,
  Divider,
  Elevation,
  Card,
  Colors,
} from '@blueprintjs/core'
import { integrationsData } from '@/data/integrations'
import {
  Providers,
  ProviderTypes,
  ProviderIcons,
  ConnectionStatus,
  ConnectionStatusLabels,
} from '@/data/Providers'
import { DataEntities, DataEntityTypes } from '@/data/DataEntities'

import ConnectionTabs from '@/components/blueprints/ConnectionTabs'
import NoData from '@/components/NoData'
import StandardStackedList from '@/components/blueprints/StandardStackedList'
import ProviderTransformationSettings from '@/components/blueprints/ProviderTransformationSettings'
import GithubSettings from '@/pages/configure/settings/github'

const DataTransformations = (props) => {
  const {
    provider,
    activeStep,
    activeConnectionTab,
    blueprintConnections = [],
    dataEntities = {},
    projects = [],
    boards = [],
    issueTypes = [],
    fields = [],
    transformations = {},
    configuredConnection,
    configuredProject,
    configuredBoard,
    handleConnectionTabChange = () => {},
    prevStep = () => {},
    addBoardTransformation = () => {},
    addProjectTransformation = () => {},
    activeTransformation,
    setTransformations = () => {},
    setTransformationSettings = () => {},
    onSave = () => {},
    onCancel = () => {},
    onClear = () => {},
    isSaving = false,
    isSavingConnection = false,
    isRunning = false,
    jiraProxyError,
    isFetchingJIRA = false
  } = props

  return (
    <div className='workflow-step workflow-step-add-transformation' data-step={activeStep?.id}>
      <p
        className='alert neutral'
      >
        Set transformation rules for your selected data to view more complex
        metrics in the dashboards.
        <br />
        <a
          href='#'
          className='more-link'
          rel='noreferrer'
          style={{
            // color: '#7497F7',
            marginTop: '5px',
            display: 'inline-block',
          }}
        >
          Find out more
        </a>
      </p>
      {blueprintConnections.length > 0 && (
        <div style={{ display: 'flex' }}>
          <div
            className='connection-tab-selector'
            style={{ minWidth: '200px' }}
          >
            <Card
              className='workflow-card connection-tabs-card'
              elevation={Elevation.TWO}
              style={{ padding: '10px' }}
            >
              <ConnectionTabs
                connections={blueprintConnections}
                onChange={handleConnectionTabChange}
                selectedTabId={activeConnectionTab}
              />
            </Card>
          </div>
          <div
            className='connection-transformation'
            style={{ marginLeft: '10px', width: '100%' }}
          >
            <Card
              className='workflow-card workflow-panel-card'
              elevation={Elevation.TWO}
            >
              {configuredConnection && (
                <>
                  <h3>
                    <span style={{ float: 'left', marginRight: '8px' }}>
                      {ProviderIcons[configuredConnection.provider]
                        ? (
                            ProviderIcons[configuredConnection.provider](24, 24)
                          )
                        : (
                          <></>
                          )}
                    </span>{' '}
                    {configuredConnection.title}
                  </h3>
                  <Divider className='section-divider' />

                  {[Providers.GITLAB, Providers.GITHUB].includes(
                    configuredConnection.provider
                  ) && (!configuredProject) && (
                    <>
                      <StandardStackedList
                        items={projects}
                        transformations={transformations}
                        className='selected-items-list selected-projects-list'
                        connection={configuredConnection}
                        activeItem={configuredProject}
                        onAdd={addProjectTransformation}
                        onChange={addProjectTransformation}
                      />
                      {projects[configuredConnection.id].length === 0 && (
                        <NoData
                          title='No Projects Selected'
                          icon='git-branch'
                          message='Please select specify at least one project.'
                          onClick={prevStep}
                        />
                      )}
                    </>
                  )}

                  {[Providers.JIRA].includes(
                    configuredConnection.provider
                  ) && (!configuredBoard) && (
                    <>
                      <StandardStackedList
                        items={boards}
                        transformations={transformations}
                        className='selected-items-list selected-boards-list'
                        connection={configuredConnection}
                        activeItem={configuredBoard}
                        onAdd={addBoardTransformation}
                        onChange={addBoardTransformation}
                      />
                      {boards[configuredConnection.id].length === 0 && (
                        <NoData
                          title='No Boards Selected'
                          icon='th'
                          message='Please select specify at least one board.'
                          onClick={prevStep}
                        />
                      )}
                    </>
                  )}

                  {(configuredProject || configuredBoard) && (
                    <div>
                      <h4>Project</h4>
                      <p style={{ color: '#292B3F' }}>{configuredProject || configuredBoard?.name || '< select a project >'}</p>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <h4 style={{ margin: 0 }}>
                          Data Transformation Rules
                        </h4>
                        <div>
                          <Button
                            minimal
                            small
                            text='Clear All'
                            intent={Intent.NONE}
                            href='#'
                            onClick={onClear}
                            style={{ float: 'right' }}
                            disabled={Object.keys(activeTransformation || {}).length === 0}
                          />
                        </div>
                      </div>

                      {!dataEntities[configuredConnection.id] ||
                        (dataEntities[configuredConnection.id]?.length ===
                          0 && <p>(No Data Entities Selected)</p>)}
                      {dataEntities[configuredConnection.id]?.find(
                        (e) => e.value === DataEntityTypes.TICKET
                      ) && (
                        <ProviderTransformationSettings
                          provider={integrationsData.find(i => i.id === configuredConnection?.provider)}
                          connection={configuredConnection}
                          configuredProject={configuredProject}
                          configuredBoard={configuredBoard}
                          issueTypes={issueTypes}
                          fields={fields}
                          boards={boards}
                          transformation={activeTransformation}
                          onSettingsChange={setTransformationSettings}
                          entity={DataEntityTypes.TICKET}
                          isSaving={isSaving}
                          isSavingConnection={isSavingConnection}
                          jiraProxyError={jiraProxyError}
                          isFetchingJIRA={isFetchingJIRA}
                        />
                      )}

                      <div className='transformation-actions' style={{ display: 'flex', justifyContent: 'flex-end' }}>
                        <Button
                          text='Cancel'
                          small
                          outlined
                          onClick={onCancel}
                        />
                        <Button
                          text='Save'
                          intent={Intent.PRIMARY}
                          small
                          outlined
                          onClick={onSave}
                          disabled={[Providers.GITLAB].includes(configuredConnection?.provider)}
                          style={{ marginLeft: '5px' }}
                        />
                      </div>
                    </div>
                  )}
                </>
              )}

              {[Providers.JENKINS].includes(configuredConnection.provider) && (
                <>
                  <div className='bp3-non-ideal-state'>
                    <div className='bp3-non-ideal-state-visual'>
                      <Icon icon='disable' size={32} />
                    </div>
                    <div className='bp3-non-ideal-state-text'>
                      <h4 className='bp3-heading' style={{ margin: 0 }}>
                        No Data Transformations
                      </h4>
                      <div>No additional settings are available at this time.</div>
                    </div>
                  </div>
                </>
              )}
            </Card>
          </div>
        </div>
      )}
      {blueprintConnections.length === 0 && (
        <>
          <div className='bp3-non-ideal-state'>
            <div className='bp3-non-ideal-state-visual'>
              <Icon icon='offline' size={32} />
            </div>
            <div className='bp3-non-ideal-state-text'>
              <h4 className='bp3-heading' style={{ margin: 0 }}>
                No Data Connections
              </h4>
              <div>Please select at least one connection source.</div>
            </div>
            <button
              className='bp3-button bp4-intent-primary'
              onClick={prevStep}
            >
              Go Back
            </button>
          </div>
        </>
      )}
    </div>
  )
}

export default DataTransformations
