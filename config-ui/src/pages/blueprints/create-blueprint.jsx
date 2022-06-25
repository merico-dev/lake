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
// import { useSelector, useDispatch } from 'react-redux'
import { CSSTransition } from 'react-transition-group'
import { useHistory, useLocation, Link } from 'react-router-dom'
import dayjs from '@/utils/time'
import {
  API_PROXY_ENDPOINT,
  ISSUE_TYPES_ENDPOINT,
  ISSUE_FIELDS_ENDPOINT,
  BOARDS_ENDPOINT,
} from '@/config/jiraApiProxy'
{/* import {
  Button,
  Icon,
  Intent,
  Switch,
  FormGroup,
  ButtonGroup,
  RadioGroup,
  Radio,
  InputGroup,
  TagInput,
  Divider,
  Elevation,
  TextArea,
  Tabs,
  Tab,
  Card,
  Popover,
  Tooltip,
  Label,
  MenuItem,
  Position,
  Colors,
  Tag,
  PopoverInteractionKind
} from '@blueprintjs/core' */}
import { integrationsData } from '@/data/integrations'
import {
  Providers,
  ProviderTypes,
  ProviderIcons,
  ConnectionStatus,
  ConnectionStatusLabels,
} from '@/data/Providers'
// import { MultiSelect, Select } from '@blueprintjs/select'
import Nav from '@/components/Nav'
import Sidebar from '@/components/Sidebar'
// import AppCrumbs from '@/components/Breadcrumbs'
import Content from '@/components/Content'

import { DataEntities, DataEntityTypes } from '@/data/DataEntities'
import { NullBlueprint } from '@/data/NullBlueprint'
import { NullBlueprintConnection } from '@/data/NullBlueprintConnection'
import { NullConnection } from '@/data/NullConnection'

import { WorkflowSteps, DEFAULT_DATA_ENTITIES, DEFAULT_BOARDS } from '@/data/BlueprintWorkflow'

import useBlueprintManager from '@/hooks/useBlueprintManager'
import usePipelineManager from '@/hooks/usePipelineManager'
import useConnectionManager from '@/hooks/useConnectionManager'
import useBlueprintValidation from '@/hooks/useBlueprintValidation'
import usePipelineValidation from '@/hooks/usePipelineValidation'
import useConnectionValidation from '@/hooks/useConnectionValidation'
import useJIRA from '@/hooks/useJIRA'

import WorkflowStepsBar from '@/components/blueprints/WorkflowStepsBar'
import WorkflowActions from '@/components/blueprints/WorkflowActions'
// import FormValidationErrors from '@/components/messages/FormValidationErrors'
// import InputValidationError from '@/components/validation/InputValidationError'
// import ConnectionsSelector from '@/components/blueprints/ConnectionsSelector'
// import DataEntitiesSelector from '@/components/blueprints/DataEntitiesSelector'
// import BoardsSelector from '@/components/blueprints/BoardsSelector'
import ConnectionDialog from '@/components/blueprints/ConnectionDialog'
// import StandardStackedList from '@/components/blueprints/StandardStackedList'
import CodeInspector from '@/components/pipelines/CodeInspector'
// import NoData from '@/components/NoData'

import DataConnections from '@/components/blueprints/create-workflow/DataConnections'
import DataScopes from '@/components/blueprints/create-workflow/DataScopes'
import DataTransformations from '@/components/blueprints/create-workflow/DataTransformations'
import DataSync from '@/components/blueprints/create-workflow/DataSync'

// import ConnectionTabs from '@/components/blueprints/ConnectionTabs'
// import ClearButton from '@/components/ClearButton'
// import CronHelp from '@/images/cron-help.png'

const CreateBlueprint = (props) => {
  const history = useHistory()
  // const dispatch = useDispatch()

  const [activeStep, setActiveStep] = useState(WorkflowSteps.find((s) => s.id === 1))
  const [advancedMode, setAdvancedMode] = useState(false)
  const [activeProvider, setActiveProvider] = useState(integrationsData[0])

  const [enabledProviders, setEnabledProviders] = useState([])
  const [runTasks, setRunTasks] = useState([])
  const [runTasksAdvanced, setRunTasksAdvanced] = useState([])
  const [existingTasks, setExistingTasks] = useState([])
  const [rawConfiguration, setRawConfiguration] = useState(
    JSON.stringify([runTasks], null, '  ')
  )
  const [isValidConfiguration, setIsValidConfiguration] = useState(false)
  const [validationError, setValidationError] = useState()

  const [connectionDialogIsOpen, setConnectionDialogIsOpen] = useState(false)
  const [managedConnection, setManagedConnection] = useState(
    NullBlueprintConnection
  )

  const [connectionsList, setConnectionsList] = useState(
    allProviderConnections?.map((c, cIdx) => ({
      ...c,
      id: cIdx,
      name: c.name,
      title: c.name,
      value: c.id,
      status:
        ConnectionStatusLabels[c.status] ||
        ConnectionStatusLabels[ConnectionStatus.OFFLINE],
      provider: c.provider,
      plugin: c.provider,
    }))
  )

  const [dataEntitiesList, setDataEntitiesList] = useState([
    ...DEFAULT_DATA_ENTITIES,
  ])
  const [boardsList, setBoardsList] = useState([...DEFAULT_BOARDS])

  const [blueprintConnections, setBlueprintConnections] = useState([])
  const [configuredConnection, setConfiguredConnection] = useState()
  const [dataEntities, setDataEntities] = useState({})
  const [activeConnectionTab, setActiveConnectionTab] = useState()

  const [showBlueprintInspector, setShowBlueprintInspector] = useState(false)

  const [dataScopes, setDataScopes] = useState([])
  const [transformations, setTransformations] = useState({})
  const [activeTransformation, setActiveTransformation] = useState()

  // @todo: replace with $projects
  const [projectId, setProjectId] = useState([])
  const [projects, setProjects] = useState({})
  const [boards, setBoards] = useState({})
  // @todo: replace with $boards
  const [boardId, setBoardId] = useState([])
  const [connectionId, setConnectionId] = useState('')
  const [connections, setConnections] = useState([])
  const [repositories, setRepositories] = useState([])
  const [selectedConnection, setSelectedConnection] = useState()
  const [repositoryName, setRepositoryName] = useState('')
  const [owner, setOwner] = useState('')
  const [gitExtractorUrl, setGitExtractorUrl] = useState('')
  const [gitExtractorRepoId, setGitExtractorRepoId] = useState('')
  const [selectedGithubRepo, setSelectedGithubRepo] = useState()
  const [refDiffRepoId, setRefDiffRepoId] = useState('')
  const [refDiffPairs, setRefDiffPairs] = useState([])
  const [refDiffTasks, setRefDiffTasks] = useState([
    'calculateCommitsDiff',
    'calculateIssuesDiff',
  ])

  const [configuredProject, setConfiguredProject] = useState(
    projects.length > 0 ? projects[0] : null
  )
  const [configuredBoard, setConfiguredBoard] = useState(
    boards.length > 0 ? boards[0] : null
  )

  const {
    activeConnection,
    // eslint-disable-next-line no-unused-vars
    blueprint,
    // eslint-disable-next-line no-unused-vars
    blueprints,
    name,
    cronConfig,
    customCronConfig,
    // eslint-disable-next-line no-unused-vars
    cronPresets,
    tasks: blueprintTasks,
    settings: blueprintSettings,
    detectedProviderTasks,
    enable,
    mode,
    setName: setBlueprintName,
    setCronConfig,
    setCustomCronConfig,
    setTasks: setBlueprintTasks,
    setSettings: setBlueprintSettings,
    setDetectedProviderTasks,
    setEnable: setEnableBlueprint,
    setMode,
    // eslint-disable-next-line no-unused-vars
    isFetching: isFetchingBlueprints,
    isSaving,
    createCronExpression: createCron,
    // eslint-disable-next-line no-unused-vars
    getCronSchedule: getSchedule,
    getNextRunDate,
    getCronPreset,
    getCronPresetByConfig,
    saveBlueprint,
    deleteBlueprint,
    isDeleting: isDeletingBlueprint,
    saveComplete: saveBlueprintComplete
  } = useBlueprintManager()

  const {
    // eslint-disable-next-line no-unused-vars
    validate: validateBlueprint,
    // eslint-disable-next-line no-unused-vars
    errors: blueprintValidationErrors,
    // setErrors: setBlueprintErrors,
    isValid: isValidBlueprint,
    fieldHasError,
    getFieldError,
  } = useBlueprintValidation({
    name,
    cronConfig,
    customCronConfig,
    enable,
    tasks: blueprintTasks,
    mode
  })

  const {
    pipelines,
    runPipeline,
    cancelPipeline,
    fetchPipeline,
    fetchAllPipelines,
    pipelineRun,
    buildPipelineStages,
    isRunning,
    isFetchingAll: isFetchingAllPipelines,
    // eslint-disable-next-line no-unused-vars
    errors: pipelineErrors,
    setSettings: setPipelineSettings,
    // eslint-disable-next-line no-unused-vars
    lastRunId,
    // eslint-disable-next-line no-unused-vars
    allowedProviders,
    // eslint-disable-next-line no-unused-vars
    detectPipelineProviders,
  } = usePipelineManager(null, runTasks)

  const {
    validate: validatePipeline,
    validateAdvanced: validateAdvancedPipeline,
    errors: validationErrors,
    setErrors: setPipelineErrors,
    isValid: isValidPipelineForm,
    detectedProviders,
  } = usePipelineValidation({
    enabledProviders,
    // pipelineName,
    projectId,
    projects,
    boardId,
    boards,
    owner,
    repositoryName,
    connectionId,
    gitExtractorUrl,
    gitExtractorRepoId,
    refDiffRepoId,
    refDiffTasks,
    refDiffPairs,
    tasks: runTasks,
    tasksAdvanced: runTasksAdvanced,
    advancedMode,
    mode,
    connection: configuredConnection,
    entities: dataEntities
  })

  const {
    fetchIssueTypes,
    fetchFields,
    issueTypes,
    fields,
    isFetching: isFetchingJIRA,
    error: jiraProxyError,
  } = useJIRA({
    apiProxyPath: API_PROXY_ENDPOINT,
    issuesEndpoint: ISSUE_TYPES_ENDPOINT,
    fieldsEndpoint: ISSUE_FIELDS_ENDPOINT,
    boardsEndpoint: BOARDS_ENDPOINT,
  })

  const {
    testConnection,
    saveConnection,
    fetchConnection,
    allProviderConnections,
    errors: connectionErrors,
    isSaving: isSavingConnection,
    isTesting: isTestingConnection,
    showError,
    testStatus,
    name: connectionName,
    endpointUrl,
    proxy,
    token,
    username,
    password,
    provider,
    setActiveConnection,
    setProvider,
    setName,
    setEndpointUrl,
    setProxy,
    setUsername,
    setPassword,
    setToken,
    fetchAllConnections,
    connectionLimitReached,
    clearConnection: clearActiveConnection
  } = useConnectionManager({
    activeProvider,
    connectionId: managedConnection?.id
  }, manageConnection?.id !== null ? true : false)

  const {
    validate: validateConnection,
    errors: connectionValidationErrors,
    isValid: isValidConnection,
  } = useConnectionValidation({
      activeProvider,
      name: connectionName,
      endpointUrl,
      proxy,
      token,
      username,
      password
  })

  const isValidStep = useCallback((stepId) => {}, [])

  const nextStep = useCallback(() => {
    setActiveStep((aS) =>
      WorkflowSteps.find((s) => s.id === Math.min(aS.id + 1, WorkflowSteps.length))
    )
  }, [WorkflowSteps])

  const prevStep = useCallback(() => {
    setActiveStep((aS) => WorkflowSteps.find((s) => s.id === Math.max(aS.id - 1, 1)))
  }, [WorkflowSteps])

  const handleConnectionTabChange = useCallback(
    (tab) => {
      console.log('>> CONNECTION TAB CHANGED', tab)
      const selectedConnection = blueprintConnections.find((c) => c.id === Number(tab.split('-')[1]))
      setActiveConnectionTab(tab)
      setActiveProvider(integrationsData.find(p => p.id === selectedConnection.provider))
      setProvider(integrationsData.find(p => p.id === selectedConnection.provider))
      setConfiguredConnection(selectedConnection)    
    },
    [blueprintConnections]
  )

  const handleConnectionDialogOpen = () => {
    console.log('>>> MANAGING CONNECTION', managedConnection)
  }

  const handleConnectionDialogClose = () => {
    setConnectionDialogIsOpen(false)
    setManagedConnection(NullBlueprintConnection)
    clearActiveConnection()
  }

  const getRestrictedDataEntities = useCallback(() => {
    let items = []
    switch (configuredConnection.provider) {
      case Providers.GITLAB:
      case Providers.JIRA:
      case Providers.GITHUB:
        items = dataEntitiesList.filter((d) => d.name !== 'ci-cd')
        break
      case Providers.JENKINS:
        items = dataEntitiesList.filter((d) => d.name == 'ci-cd')
        break
        return items
    }
  }, [dataEntitiesList, configuredConnection])

  const createProviderScopes = useCallback(
    (
      providerId,
      connection,
      connectionIdx,
      entities = [],
      boards = [],
      projects = [],
      transformations = [],
      defaultScope = { transformation: {}, options: {}, entities: [] }
    ) => {
      console.log(
        '>>> CREATING PROVIDER SCOPE FOR CONNECTION...',
        connectionIdx,
        connection
      )
      let newScope = {
        ...defaultScope,
        entities: entities[connection.id]?.map((entity) => entity.value) || [],
      }
      switch (providerId) {
        case Providers.JIRA:
          newScope = boards[connection.id]?.map((b) => ({
            ...newScope,
            options: {
              boardId: b.id,
              // @todo: verify initial value of since date for jira provider
              since: new Date(),
            },
            // @todo: verify transformation payload for jira
            transformation: {},
          }))
          break
        case Providers.GITLAB:
          newScope = projects[connection.id]?.map((p) => ({
            ...newScope,
            options: {
              projectId: p,
            },
            // @todo: verify transformation payload for gitlab (none? - no additional settings)
            transformation: {},
          }))
          break
        case Providers.JENKINS:
          // @todo: verify scope settings if any for jenkins
          newScope = {
            ...newScope,
          }
          break
        case Providers.GITHUB:
          newScope = projects[connection.id]?.map((p) => ({
            ...newScope,
            options: {
              owner: p.split('/')[0],
              repo: p.split('/')[1],
            },
            transformation: { ...transformations[p] },
          }))
          break
      }
      return Array.isArray(newScope) ? newScope.flat() : [newScope]
    },
    []
  )

  const manageConnection = useCallback((connection) => {
    console.log('>> MANAGE CONNECTION...', connection)
    if (connection?.id !== null) {
      setActiveProvider(integrationsData.find(p => p.id === connection.provider))
      setProvider(integrationsData.find(p => p.id === connection.provider))
      // fetchConnection(true, false, connection.id)
      setManagedConnection(connection)
      setConnectionDialogIsOpen(true)
    }
  }, [])

  const addProjectTransformation = (project) => {
    setConfiguredProject(project)
  }

  const addBoardTransformation = (board) => {
    setConfiguredBoard(board)
  }
  
  const addConnection = () => {
    setManagedConnection(NullBlueprintConnection)
    setConnectionDialogIsOpen(true)
  }
  
  const setTransformationSettings = useCallback((settings, configuredProject) => {
    // @todo: fix configuredProject is null here!
    console.log(`>> SETTING TRANSFORMATION SETTINGS [PROJECT = ${configuredProject}]...`, settings)
    setTransformations(existingTransformations => ({
      ...existingTransformations,
      [configuredProject]: { ...settings }
    }))
  }, [])
  

  useEffect(() => {
    console.log('>> ACTIVE STEP CHANGED: ', activeStep)
    if (activeStep?.id === 1) {
      const enableNotifications = false
      const getAllSources = true
      fetchAllConnections(enableNotifications, getAllSources)
    }
  }, [activeStep])

  useEffect(() => {
    console.log('>>> ALL DATA PROVIDER CONNECTIONS...', allProviderConnections)
    setConnectionsList(
      allProviderConnections?.map((c, cIdx) => ({
        ...c,
        id: cIdx,
        name: c.name,
        title: c.name,
        value: c.id,
        status:
          ConnectionStatusLabels[c.status] ||
          ConnectionStatusLabels[ConnectionStatus.OFFLINE],
        provider: c.provider,
        plugin: c.provider,
      }))
    )
  }, [allProviderConnections])

  useEffect(() => {
    console.log(
      '>> PIPELINE RUN TASK SETTINGS FOR PIPELINE MANAGER ....',
      runTasks
    )
    // setPipelineSettings({
    //   name: pipelineName,
    //   tasks: advancedMode ? runTasksAdvanced : [[...runTasks]]
    // })
    // setRawConfiguration(JSON.stringify(buildPipelineStages(runTasks, true), null, '  '))
    if (advancedMode) {
      validateAdvancedPipeline()
      setBlueprintTasks(runTasksAdvanced)
    } else {
      validatePipeline()
      setBlueprintTasks([[...runTasks]])
    }
  }, [
    advancedMode,
    runTasks,
    runTasksAdvanced,
    setPipelineSettings,
    validatePipeline,
    validateAdvancedPipeline,
    setBlueprintTasks,
  ])

  useEffect(() => {
    console.log(
      '>> BLUEPRINT SETTINGS FOR PIPELINE MANAGER ....',
      blueprintSettings
    )
  }, [blueprintSettings])

  useEffect(() => {
    validateBlueprint()
  }, [
    name,
    cronConfig,
    customCronConfig,
    blueprintTasks,
    enable,
    validateBlueprint,
  ])

  useEffect(() => {}, [activeConnectionTab])

  useEffect(() => {
    setConfiguredConnection(
      blueprintConnections.length > 0 ? blueprintConnections[0] : null
    )
    const initializeEntities = (pV, cV) => ({ ...pV, [cV.id]: [] })
    const initializeProjects = (pV, cV) => ({ ...pV, [cV.id]: [] })
    const initializeBoards = (pV, cV) => ({ ...pV, [cV.id]: [] })
    setDataEntities((dE) => ({
      ...blueprintConnections.reduce(initializeEntities, {}),
    }))
    setProjects((p) => ({
      ...blueprintConnections.reduce(initializeProjects, {}),
    }))
    setBoards((b) => ({
      ...blueprintConnections.reduce(initializeBoards, {}),
    }))
    setEnabledProviders([
      ...new Set(blueprintConnections.map((c) => c.provider)),
    ])
  }, [blueprintConnections])

  useEffect(() => {
    console.log('>> CONFIGURING CONNECTION', configuredConnection)
    if (configuredConnection) {
      switch (configuredConnection.provider) {
        case Providers.GITLAB:
        case Providers.JIRA:
        case Providers.GITHUB:
          setDataEntitiesList(
            DEFAULT_DATA_ENTITIES.filter((d) => d.name !== 'ci-cd')
          )
          // setConfiguredProject(projects.length > 0 ? projects[0] : null)
          break
        case Providers.JENKINS:
          setDataEntitiesList(
            DEFAULT_DATA_ENTITIES.filter((d) => d.name == 'ci-cd')
          )
          break
        default:
          setDataEntitiesList(DEFAULT_DATA_ENTITIES)
          break
      }
    }
  }, [configuredConnection])

  useEffect(() => {
    console.log('>> DATA ENTITIES', dataEntities)
  }, [dataEntities])

  useEffect(() => {
    console.log('>> BOARDS', boards)
  }, [boards])

  useEffect(() => {
    setBlueprintSettings((currentSettings) => ({
      ...currentSettings,
      connections: blueprintConnections.map((c, cIdx) => ({
        ...NullBlueprintConnection,
        connectionId: c.value,
        plugin: c.plugin || c.provider,
        scope: createProviderScopes(
          c.provider,
          c,
          cIdx,
          dataEntities,
          boards,
          projects,
          transformations
        ),
      })),
    }))
    // validatePipeline()
  }, [blueprintConnections, dataEntities, boards, projects, transformations, validatePipeline])

  useEffect(() => {
    console.log('>> PROJECTS LIST', projects)
    const getDefaultTransformations = (providerId) => {
      let transforms = {}
        switch(providerId) {
          case Providers.GITHUB:
            transforms = {
              prType: '',
              prComponent: '',
              issueSeverity: '',
              issueComponent: '',
              issuePriority: '',
              issueTypeRequirement: '',
              issueTypeBug: '',
              issueTypeIncident: '',
            }
          break
          case Providers.JIRA:
          break
          case Providers.JENKINS:
          break
          case Providers.GITLAB:
          break
        }
      return transforms
    }
    // @todo: check if this setter is required at this level
    // @todo: fix lost transformations when connection tabs switch
    // setConfiguredProject(projects.length > 0 ? projects[0] : null)
    const initializeTransformations = (pV, cV) => ({ ...pV, [cV]: getDefaultTransformations(configuredConnection?.provider)})
    const projectTransformation = projects[configuredConnection?.id]
    if (projectTransformation) {
      setTransformations(cT => ({
        ...projectTransformation.reduce(initializeTransformations, {})
      }))
    }
  }, [projects, configuredConnection])

  useEffect(() => {
    console.log('>>> SELECTED PROJECT TO CONFIGURE...', configuredProject)
    setActiveTransformation(aT => configuredProject ? transformations[configuredProject] : aT)
  }, [configuredProject, transformations])

  useEffect(() => {
    console.log('>>> ACTIVE/MODIFYING TRANSFORMATION RULES...', activeTransformation)
  }, [activeTransformation])

  return (
    <>
      <div className='container'>
        <Nav />
        <Sidebar />
        <Content>
          <main className='main'>
            <WorkflowStepsBar activeStep={activeStep} />

            <div
              className={`workflow-content workflow-step-id-${activeStep?.id}`}
            >
              {activeStep?.id === 1 && (
                <DataConnections
                  activeStep={activeStep}
                  blueprintConnections={blueprintConnections}
                  connectionsList={connectionsList}
                  setBlueprintName={setBlueprintName}
                  setBlueprintConnections={setBlueprintConnections}
                  fieldHasError={fieldHasError}
                  getFieldError={getFieldError}
                  addConnection={addConnection}
                  manageConnection={manageConnection}
                />
              )}

              {activeStep?.id === 2 && (
                <DataScopes
                  provider={provider}
                  activeStep={activeStep}
                  activeConnectionTab={activeConnectionTab}
                  blueprintConnections={blueprintConnections}
                  dataEntitiesList={dataEntitiesList}
                  boardsList={boardsList}
                  boards={boards}
                  dataEntities={dataEntities}
                  projects={projects}
                  configuredConnection={configuredConnection}
                  handleConnectionTabChange={handleConnectionTabChange}
                  setDataEntities={setDataEntities}
                  setProjects={setProjects}
                  setBoards={setBoards}
                  prevStep={prevStep}
                  isSaving={isSaving}
                  isRunning={isRunning}
                />
              )}

              {activeStep?.id === 3 && (
                <DataTransformations
                  provider={provider}
                  activeStep={activeStep}
                  activeConnectionTab={activeConnectionTab}
                  blueprintConnections={blueprintConnections}
                  dataEntities={dataEntities}
                  projects={projects}
                  boards={boards}
                  configuredConnection={configuredConnection}
                  configuredProject={configuredProject}
                  configurdBoard={configuredBoard}
                  handleConnectionTabChange={handleConnectionTabChange}
                  prevStep={prevStep}
                  addBoardTransformation={addBoardTransformation}
                  addProjectTransformation={addProjectTransformation}
                  transformations={transformations}
                  activeTransformation={activeTransformation}
                  setTransformations={setTransformations}
                  setTransformationSettings={setTransformationSettings}
                  isSaving={isSaving}
                  isSavingConnection={isSavingConnection}
                  isRunning={isRunning}
                />
              )}

              {activeStep?.id === 4 && (
                <DataSync
                  cronConfig={cronConfig}
                  customCronConfig={customCronConfig}
                  createCron={createCron}
                  setCronConfig={setCronConfig}
                  getCronPreset={getCronPreset}
                  fieldHasError={fieldHasError}
                  getFieldError={getFieldError}
                  setCustomCronConfig={setCustomCronConfig}
                  getCronPresetByConfig={getCronPresetByConfig}
                />
              )}
            </div>

            <WorkflowActions 
              activeStep={activeStep}
              setShowBlueprintInspector={setShowBlueprintInspector}
              validationErrors={validationErrors}
              nextStep={nextStep}
              prevStep={prevStep}
            />
          </main>
        </Content>
      </div>
      
      <ConnectionDialog
        integrations={integrationsData}
        activeProvider={activeProvider}
        setProvider={setActiveProvider}
        connection={managedConnection}
        errors={connectionErrors}
        validationErrors={connectionValidationErrors}
        endpointUrl={endpointUrl}
        name={connectionName}
        proxy={proxy}
        token={token}
        username={username}
        password={password}
        isOpen={connectionDialogIsOpen}
        isTesting={isTestingConnection}
        isSaving={isSavingConnection}
        isValid={isValidConnection}
        onClose={handleConnectionDialogClose}
        onOpen={handleConnectionDialogOpen}
        onTest={testConnection}
        onSave={saveConnection}
        onValidate={validateConnection}
        onNameChange={setName}
        onEndpointChange={setEndpointUrl}
        onProxyChange={setProxy}
        onTokenChange={setToken}
        onUsernameChange={setUsername}
        onPasswordChange={setPassword}
        testStatus={testStatus}
      />

      <CodeInspector
        title={name}
        titleIcon='add'
        subtitle='JSON CONFIGURATION'
        isOpen={showBlueprintInspector}
        activePipeline={{
          ID: 0,
          name,
          plan: blueprintTasks,
          settings: blueprintSettings,
          cronConfig,
          enable,
        }}
        onClose={setShowBlueprintInspector}
        hasBackdrop={false}
      />
    </>
  )
}

export default CreateBlueprint
