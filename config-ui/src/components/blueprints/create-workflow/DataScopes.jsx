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
  TagInput,
  Divider,
  Elevation,
  Card,
  Colors,
} from '@blueprintjs/core'
import {
  Providers,
  ProviderTypes,
  ProviderIcons,
  ConnectionStatus,
  ConnectionStatusLabels,
} from '@/data/Providers'

import ConnectionTabs from '@/components/blueprints/ConnectionTabs'
import BoardsSelector from '@/components/blueprints/BoardsSelector'
import DataEntitiesSelector from '@/components/blueprints/DataEntitiesSelector'
import NoData from '@/components/NoData'

const DataScopes = (props) => {
  const {
    activeStep,
    activeConnectionTab,
    blueprintConnections = [],
    dataEntitiesList = [],
    boardsList = [],
    dataEntities = [],
    projects = [],
    boards = [],
    configuredConnection,
    handleConnectionTabChange = () => {},
    setDataEntities = () => {},
    setProjects = () => {},
    setBoards = () => {},
    prevStep = () => {},
    isSaving = false,
    isRunning = false,
  } = props

  return (
    <div className='workflow-step workflow-step-data-scope' data-step={activeStep?.id}>
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
            className='connection-scope'
            style={{ marginLeft: '10px', width: '100%' }}
          >
            <Card
              className='workflow-card worfklow-panel-card'
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
                  ) && (
                    <>
                      <h4>Projects *</h4>
                      <p>Enter the projects you would like to sync.</p>
                      <TagInput
                        id='project-id'
                        disabled={isRunning}
                        placeholder={
                          configuredConnection.provider === Providers.GITHUB
                            ? 'username/repo, username/another-repo'
                            : '1000000, 200000'
                        }
                        values={projects[configuredConnection.id] || []}
                        fill={true}
                        onChange={(values) =>
                          setProjects((p) => ({
                            ...p,
                            [configuredConnection.id]: [...new Set(values)],
                          }))}
                        addOnPaste={true}
                        addOnBlur={true}
                        rightElement={
                          <Button
                            disabled={isRunning}
                            icon='eraser'
                            minimal
                            onClick={() =>
                              setProjects((p) => ({
                                ...p,
                                [configuredConnection.id]: [],
                              }))}
                          />
                        }
                        onKeyDown={(e) =>
                          e.key === 'Enter' && e.preventDefault()}
                        tagProps={{
                          intent: Intent.PRIMARY,
                          minimal: true,
                        }}
                        className='input-project-id tagInput'
                      />
                    </>
                  )}

                  {[Providers.JIRA].includes(configuredConnection.provider) && (
                    <>
                      <h4>Boards *</h4>
                      <p>Select the boards you would like to sync.</p>
                      <BoardsSelector
                        items={boardsList}
                        selectedItems={boards[configuredConnection.id] || []}
                        onItemSelect={setBoards}
                        onClear={setBoards}
                        onRemove={setBoards}
                        disabled={isSaving}
                        configuredConnection={configuredConnection}
                      />
                    </>
                  )}

                  <h4>Data Entities</h4>
                  <p>
                    Select the data entities you wish to collect for the
                    projects.{' '}
                    <a
                      href='https://devlake.apache.org/docs/DataModels/DevLakeDomainLayerSchema'
                      target='_blank'
                      rel='noreferrer'
                    >
                      Learn about data entities
                    </a>
                  </p>
                  <DataEntitiesSelector
                    items={dataEntitiesList}
                    selectedItems={dataEntities[configuredConnection.id] || []}
                    // restrictedItems={getRestrictedDataEntities()}
                    onItemSelect={setDataEntities}
                    onClear={setDataEntities}
                    onRemove={setDataEntities}
                    disabled={isSaving}
                    configuredConnection={configuredConnection}
                    isSaving={isSaving}
                  />
                </>
              )}
            </Card>
          </div>
        </div>
      )}
      {blueprintConnections.length === 0 && (
        <NoData
          title='No Data Connections'
          message='Please select at least one connection source.'
          onClick={prevStep}
        />
      )}
    </div>
  )
}

export default DataScopes
