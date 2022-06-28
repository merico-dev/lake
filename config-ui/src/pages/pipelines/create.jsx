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
import React, { Fragment, useEffect, useCallback, useState, useRef } from 'react'
import { CSSTransition } from 'react-transition-group'
import {
  useHistory,
  useLocation,
  Link,
} from 'react-router-dom'
import { GRAFANA_URL } from '@/utils/config'
import {
  Button, Icon, Intent, Switch,
  FormGroup, ButtonGroup, InputGroup,
  Elevation,
  TextArea,
  Card,
  Popover,
  Tooltip,
  Position,
  Colors,
  Tag
} from '@blueprintjs/core'
import {
  Providers,
  ProviderTypes,
  ProviderIcons
} from '@/data/Providers'
import { integrationsData, pluginsData } from '@/data/integrations'
import useBlueprintManager from '@/hooks/useBlueprintManager'
import usePipelineManager from '@/hooks/usePipelineManager'
import useBlueprintValidation from '@/hooks/useBlueprintValidation'
import usePipelineValidation from '@/hooks/usePipelineValidation'
import useConnectionManager from '@/hooks/useConnectionManager'
import FormValidationErrors from '@/components/messages/FormValidationErrors'
import PipelineIndicator from '@/components/widgets/PipelineIndicator'
import PipelinePresetsMenu from '@/components/menus/PipelinePresetsMenu'
import PipelineConfigsMenu from '@/components/menus/PipelineConfigsMenu'
import ProviderSettings from '@/components/pipelines/ProviderSettings'
import Nav from '@/components/Nav'
import Sidebar from '@/components/Sidebar'
import AppCrumbs from '@/components/Breadcrumbs'
import Content from '@/components/Content'
import AddBlueprintDialog from '@/components/blueprints/AddBlueprintDialog'
// import CodeEditor from '@uiw/react-textarea-code-editor'
import { ReactComponent as HelpIcon } from '@/images/help.svg'

import GitlabHelpNote from '@/images/help/gitlab-help.png'
import JiraHelpNote from '@/images/help/jira-help.png'
import GithubHelpNote from '@/images/help/github-help.png'

import '@/styles/pipelines.scss'

const CreatePipeline = (props) => {
  const history = useHistory()
  const location = useLocation()
  // const { providerId } = useParams()
  // const [activeProvider, setActiveProvider] = useState(integrationsData[0])
  // eslint-disable-next-line no-unused-vars
  const [integrations, setIntegrations] = useState([...integrationsData, ...pluginsData])
  // eslint-disable-next-line no-unused-vars
  const [jiraIntegration, setJiraIntegration] = useState(integrationsData.find(p => p.id === Providers.JIRA))

  const [today, setToday] = useState(new Date())
  const pipelinePrefixes = ['COLLECT', 'SYNC']
  const pipelineSuffixes = [
    today.getTime(), // 1639630123107
    today.toString(), // Wed Dec 15 2021 23:48:43 GMT-0500 (EST)
    today.toISOString(), // 2021-12-16T04:48:43.107Z
    // eslint-disable-next-line max-len
    `${today.getFullYear()}${today.getMonth() + 1}${today.getDate()}${today.getHours()}${today.getMinutes()}${today.getSeconds()}`, // 202112154936
    today.toUTCString(), // Thu, 16 Dec 2021 04:49:52 GMT
  ]

  const [readyProviders, setReadyProviders] = useState([])
  const [advancedMode, setAdvancedMode] = useState(false)
  const [enableAutomation, setEnableAutomation] = useState(false)
  const [blueprintDialogIsOpen, setBlueprintDialogIsOpen] = useState(false)
  const [draftBlueprint, setDraftBlueprint] = useState(null)
  const [pipelineTemplates, setPipelineTemplates] = useState([])
  const [selectedPipelineTemplate, setSelectedPipelineTemplate] = useState()

  const [enabledProviders, setEnabledProviders] = useState([])
  const [runTasks, setRunTasks] = useState([])
  const [runTasksAdvanced, setRunTasksAdvanced] = useState([])
  const [existingTasks, setExistingTasks] = useState([])
  const [rawConfiguration, setRawConfiguration] = useState(JSON.stringify([runTasks], null, '  '))
  const [isValidConfiguration, setIsValidConfiguration] = useState(false)
  const [validationError, setValidationError] = useState()

  const [namePrefix, setNamePrefix] = useState(pipelinePrefixes[0])
  const [nameSuffix, setNameSuffix] = useState(pipelineSuffixes[0])
  const [pipelineName, setPipelineName] = useState(`${namePrefix} ${nameSuffix}`)
  const [projectId, setProjectId] = useState([])
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
  const [refDiffTasks, setRefDiffTasks] = useState(['calculateCommitsDiff', 'calculateIssuesDiff'])

  const addBlueprintRef = useRef()

  // eslint-disable-next-line no-unused-vars
  const [autoRedirect, setAutoRedirect] = useState(true)
  // eslint-disable-next-line no-unused-vars
  const [restartDetected, setRestartDetected] = useState(false)

  const {
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
    detectedProviderTasks,
    enable,
    setName: setBlueprintName,
    setCronConfig,
    setCustomCronConfig,
    setTasks: setBlueprintTasks,
    setDetectedProviderTasks,
    setEnable: setEnableBlueprint,
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
    saveComplete: saveBlueprintComplete,
  } = useBlueprintManager()

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
    detectPipelineProviders
  } = usePipelineManager(pipelineName, runTasks)

  const {
    // eslint-disable-next-line no-unused-vars
    validate: validateBlueprint,
    // eslint-disable-next-line no-unused-vars
    errors: blueprintValidationErrors,
    // setErrors: setBlueprintErrors,
    isValid: isValidBlueprint,
    fieldHasError,
    getFieldError
  } = useBlueprintValidation({
    name,
    cronConfig,
    customCronConfig,
    enable,
    tasks: blueprintTasks
  })

  const {
    validate,
    validateAdvanced,
    errors: validationErrors,
    setErrors: setPipelineErrors,
    isValid: isValidPipelineForm,
    detectedProviders
  } = usePipelineValidation({
    enabledProviders,
    pipelineName,
    projectId,
    boardId,
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
    advancedMode
  })

  const {
    allConnections,
    domainRepositories,
    // eslint-disable-next-line no-unused-vars
    isFetching: isFetchingConnections,
    fetchAllConnections,
    fetchDomainLayerRepositories,
    // eslint-disable-next-line no-unused-vars
    getConnectionName
  } = useConnectionManager({
    activeProvider: jiraIntegration
  })

  useEffect(() => {
    [...integrationsData, ...pluginsData].forEach((i, idx) => {
      setTimeout(() => {
        setReadyProviders(r => [...r, i.id])
      }, idx * 50)
    })
  }, [])

  const isProviderEnabled = (providerId) => {
    return enabledProviders.includes(providerId)
  }

  const isValidPipeline = () => {
    if (advancedMode) { return isValidAdvancedPipeline() }
    return enabledProviders.length >= 1 &&
      pipelineName !== '' &&
      pipelineName.length > 2 &&
      validationErrors.length === 0
  }

  const isValidAdvancedPipeline = () => {
    return pipelineName !== '' &&
    pipelineName.length > 2 &&
    validationErrors.length === 0 &&
    isValidConfiguration
  }

  const isMultiStagePipeline = (tasks = []) => {
    return tasks.length > 1 && Array.isArray(tasks[0])
  }

  const getManyProviderOptions = useCallback((providerId, optionProperty, ids, options = {}) => {
    return ids.map(id => {
      return {
        Plugin: providerId,
        Options: {
          [optionProperty]: parseInt(id, 10),
          ...options
        }
      }
    })
  }, [])

  const getProviderOptions = useCallback((providerId) => {
    let options = {}
    switch (providerId) {
      case Providers.JENKINS:
        // NO OPTIONS for Jenkins!
        break
      case Providers.JIRA:
        options = {
          boardId: parseInt(boardId, 10),
          connectionId: parseInt(connectionId, 10)
        }
        break
      case Providers.GITHUB:
        options = {
          repo: repositoryName,
          owner
        }
        break
      case Providers.GITLAB:
        options = {
          projectId: parseInt(projectId, 10)
        }
        break
      case Providers.GITEXTRACTOR:
        options = {
          url: gitExtractorUrl,
          repoId: gitExtractorRepoId
        }
        break
      case Providers.REFDIFF:
        options = {
          repoId: refDiffRepoId,
          pairs: refDiffPairs,
          tasks: refDiffTasks,
        }
        break
      default:
        break
    }
    return options
  }, [boardId,
    owner,
    projectId,
    repositoryName,
    connectionId,
    gitExtractorUrl,
    gitExtractorRepoId,
    refDiffRepoId,
    refDiffTasks,
    refDiffPairs
  ])

  const configureProvider = useCallback((providerId) => {
    let providerConfig = {}
    switch (providerId) {
      case Providers.GITLAB:
        providerConfig = getManyProviderOptions(providerId, 'projectId', [...projectId])
        break
      case Providers.JIRA:
        providerConfig = getManyProviderOptions(
          providerId,
          'boardId',
          [...boardId],
          {
            connectionId: parseInt(connectionId, 10)
          }
        )
        break
      default:
        providerConfig = {
          Plugin: providerId,
          Options: {
            ...getProviderOptions(providerId)
          }
        }
        break
    }
    return providerConfig
  }, [getProviderOptions, getManyProviderOptions, projectId, boardId, connectionId])

  const resetPipelineName = () => {
    setToday(new Date())
    setPipelineName(`${namePrefix} ${nameSuffix}`)
  }

  const resetConfiguration = () => {
    window.history.replaceState(null, '')
    resetPipelineName()
    setExistingTasks([])
    setEnabledProviders([])
    setProjectId([])
    setBoardId([])
    setSelectedConnection(null)
    setRepositoryName('')
    setOwner('')
    setGitExtractorUrl('')
    setGitExtractorRepoId('')
    setRefDiffRepoId('')
    setRefDiffTasks([])
    setRefDiffPairs([])
    setAdvancedMode(false)
    setRawConfiguration('[[]]')
  }

  const parseJSON = (jsonString = '') => {
    try {
      return JSON.parse(jsonString)
    } catch (e) {
      console.log('>> PARSE JSON ERROR!', e)
      setValidationError(e.message)
      setPipelineErrors(errs => [...errs, e.message])
      // ToastNotification.show({ message: e.message, intent: 'danger', icon: 'error' })
    }
  }

  const formatRawCode = () => {
    try {
      setRawConfiguration(config => {
        const parsedConfig = parseJSON(config)
        const formattedConfig = JSON.stringify(parsedConfig, null, '  ')
        return formattedConfig || config
      })
    } catch (e) {
      console.log('>> FORMAT CODE: Invalid Code Format!')
    }
  }

  const isValidCode = useCallback(() => {
    let isValid = false
    try {
      const parsedCode = parseJSON(rawConfiguration)
      isValid = parsedCode
      // setValidationError(null)
    } catch (e) {
      console.log('>> FORMAT CODE: Invalid Code Format!', e)
      setValidationError(e.message)
    }
    setIsValidConfiguration(isValid)
    return isValid
  }, [rawConfiguration])

  useEffect(() => {

  }, [pipelineName])

  useEffect(() => {
    console.log('>> PIPELINE RUN TASK SETTINGS FOR PIPELINE MANAGER ....', runTasks)
    setPipelineSettings({
      name: pipelineName,
      tasks: advancedMode ? runTasksAdvanced : [[...runTasks]]
    })
    // setRawConfiguration(JSON.stringify(buildPipelineStages(runTasks, true), null, '  '))
    if (advancedMode) {
      validateAdvanced()
      setBlueprintTasks(runTasksAdvanced)
    } else {
      validate()
      setBlueprintTasks([[...runTasks]])
    }
  }, [advancedMode, runTasks, runTasksAdvanced, pipelineName, setPipelineSettings, validate, validateAdvanced, setBlueprintTasks])

  useEffect(() => {
    validateBlueprint()
  }, [name, cronConfig, customCronConfig, blueprintTasks, enable, validateBlueprint])

  useEffect(() => {
    console.log('>> ENBALED PROVIDERS = ', enabledProviders)
    const PipelineTasks = enabledProviders.map(p => configureProvider(p))
    setRunTasks(PipelineTasks.flat())
    console.log('>> CONFIGURED PIPELINE TASKS = ', PipelineTasks)
    validate()
    if (enabledProviders.includes(Providers.JIRA)) {
      fetchAllConnections(false)
    } else {
      setConnections([])
      setSelectedConnection(null)
    }
    if (enabledProviders.includes(Providers.GITEXTRACTOR)) {
      fetchDomainLayerRepositories()
    } else {
      setRepositories([])
      setSelectedGithubRepo(null)
    }
  }, [
    enabledProviders,
    projectId,
    boardId, connectionId,
    owner, repositoryName,
    configureProvider,
    validate,
    fetchAllConnections,
    fetchDomainLayerRepositories,
    buildPipelineStages
  ])

  useEffect(() => {
    console.log('>> PIPELINE LAST RUN OBJECT CHANGED!!...', pipelineRun)
    if (pipelineRun.ID && autoRedirect) {
      history.push(`/pipelines/activity/${pipelineRun.ID}`)
    }
  }, [pipelineRun, autoRedirect, history])

  useEffect(() => {
    console.log(namePrefix, nameSuffix)
    setPipelineName(`${namePrefix} ${nameSuffix}`)
    setToday(new Date())
  }, [namePrefix, nameSuffix])

  useEffect(() => {
    console.log('>> JIRA CONNECTION ID SELECTED, CONNECTION INSTANCE = ', selectedConnection)
    setConnectionId(sId => selectedConnection ? selectedConnection.value : null)
    validate()
  }, [selectedConnection, validate])

  useEffect(() => {
    console.log('>> DOMAIN LAYER REPOSITIRY SELECTED, REPO = ', selectedGithubRepo)
    setGitExtractorRepoId(rId => selectedGithubRepo ? selectedGithubRepo.value : null)
    validate()
  }, [selectedGithubRepo, validate])

  useEffect(() => {
    console.log('>> FETCHED ALL JIRA CONNECTIONS... ', allConnections)
    setConnections(allConnections.map(c => { return { id: c.ID, title: c.name || 'Instance', value: c.ID } }))
  }, [allConnections])

  useEffect(() => {
    console.log('>> FETCHED DOMAIN LAYER REPOS... ', domainRepositories)
    setRepositories(domainRepositories.map((r, rIdx) => { return { id: rIdx, title: r.name || r.id || `Repository #${r.id || rIdx}`, value: r.id || rIdx } }))
  }, [domainRepositories])

  useEffect(() => {
    console.log('>> BUILT JIRA INSTANCE SELECT MENU... ', connections)
  }, [connections])

  useEffect(() => {
    if (location.state?.existingTasks) {
      console.log('>> RESTART ATTEMPT: DETECTED EXISTING PIPELINE CONFIGURATION... ', location.state.existingTasks)
      const tasks = location.state.existingTasks
      setRestartDetected(true)
      setExistingTasks(tasks)
      window.history.replaceState(null, '')
      // !WARNING! This logic will only handle ONE STAGE (Stage 1)
      // @todo: refactor later for multi-stage
      const GitLabTask = tasks.filter(t => t.plugin === Providers.GITLAB)
      const GitHubTask = tasks.find(t => t.plugin === Providers.GITHUB)
      const JiraTask = tasks.filter(t => t.plugin === Providers.JIRA)
      const JenkinsTask = tasks.find(t => t.plugin === Providers.JENKINS)
      const GitExtractorTask = tasks.find(t => t.plugin === Providers.GITEXTRACTOR)
      const RefDiffTask = tasks.find(t => t.plugin === Providers.REFDIFF)
      const configuredProviders = []
      if (GitLabTask && GitLabTask.length > 0) {
        configuredProviders.push(Providers.GITLAB)
        setProjectId(Array.isArray(GitLabTask) ? GitLabTask.map(gT => gT.options?.projectId) : GitLabTask.options?.projectId)
      }
      if (GitHubTask) {
        configuredProviders.push(Providers.GITHUB)
        setRepositoryName(GitHubTask.options?.repositoryName || GitHubTask.options?.repo)
        setOwner(GitHubTask.options?.owner)
      }
      if (JiraTask && JiraTask.length > 0) {
        fetchAllConnections(false)
        configuredProviders.push(Providers.JIRA)
        setBoardId(Array.isArray(JiraTask) ? JiraTask.map(jT => jT.options?.boardId) : JiraTask.options?.boardId)
        const connSrcId = JiraTask[0].options?.connectionId
        setSelectedConnection({
          id: parseInt(connSrcId, 10),
          title: '(Instance)',
          value: parseInt(connSrcId, 10)
        })
      }
      if (JenkinsTask) {
        configuredProviders.push(Providers.JENKINS)
      }
      if (GitExtractorTask) {
        setGitExtractorRepoId(GitExtractorTask.options?.repoId)
        setGitExtractorUrl(GitExtractorTask.options?.url)
        configuredProviders.push(Providers.GITEXTRACTOR)
      }
      if (RefDiffTask) {
        setRefDiffRepoId(RefDiffTask.options?.repoId)
        setRefDiffTasks(RefDiffTask.options?.tasks || [])
        setRefDiffPairs(RefDiffTask.options?.pairs || [])
        configuredProviders.push(Providers.REFDIFF)
      }
      setEnabledProviders(eP => [...eP, ...configuredProviders])
    } else {
      setRestartDetected(false)
      setExistingTasks([])
    }

    return () => {
      setRestartDetected(false)
      setExistingTasks([])
    }
  }, [location, fetchAllConnections])

  useEffect(() => {
    if (isValidCode()) {
      setRunTasksAdvanced(JSON.parse(rawConfiguration))
    }
  }, [rawConfiguration, isValidCode])

  useEffect(() => {
    if (existingTasks.length > 0) {
      const multiStageTasks = buildPipelineStages(existingTasks, true)
      const PipelineTasks = multiStageTasks.map(s => {
        return s.map(t => {
          return {
            Plugin: t.plugin,
            Options: {
              ...t.options
            }
          }
        })
      })
      setRunTasksAdvanced(PipelineTasks)
      setRawConfiguration(JSON.stringify(PipelineTasks, null, '  '))
    }
  }, [existingTasks, buildPipelineStages])

  useEffect(() => {
    console.log('>>> ADVANCED MODE ENABLED?: ', advancedMode)
  }, [advancedMode])

  useEffect(() => {
    if (blueprintDialogIsOpen) {
      fetchAllPipelines('TASK_COMPLETED', 100)
    }
  }, [blueprintDialogIsOpen, fetchAllPipelines])

  useEffect(() => {
    setPipelineTemplates(pipelines.slice(0, 100).map(p => ({ ...p, id: p.id, title: p.name, value: p.id })))
  }, [pipelines])

  useEffect(() => {
    if (selectedPipelineTemplate) {
      // !! DISABLED FOR CREATE MODE !!
      // setBlueprintTasks(selectedPipelineTemplate.tasks)
    }
  }, [selectedPipelineTemplate])

  useEffect(() => {
    setSelectedPipelineTemplate(pipelineTemplates.find(pT => pT.tasks.flat().toString() === blueprintTasks.flat().toString()))
  }, [pipelineTemplates])

  useEffect(() => {
    if (saveBlueprintComplete && saveBlueprintComplete?.id) {
      setDraftBlueprint(saveBlueprintComplete)
      setBlueprintDialogIsOpen(false)
    }
  }, [saveBlueprintComplete])

  useEffect(() => {
    if (Array.isArray(blueprintTasks)) {
      setDetectedProviderTasks(blueprintTasks.flat())
    }
  }, [blueprintTasks, setDetectedProviderTasks])

  useEffect(() => {
    console.log('>>>> DETECTED PROVIDERS TASKS....', detectedProviderTasks)
  }, [detectedProviderTasks])

  useEffect(() => {
    if (enableAutomation && !blueprintDialogIsOpen) {
      if (addBlueprintRef) {
        addBlueprintRef.current?.buttonRef.click()
      }
    }
    // NOTE: do NOT include $blueprintDialogIsOpen to deps -- excluded intentionally!
    // This will allow auto-open to fire only once when automation swtich is toggled.
  }, [enableAutomation])

  return (
    <>
      <div className={`container container-create-pipeline ${advancedMode ? 'advanced-mode' : ''}`}>
        <Nav />
        <Sidebar />
        <Content>
          <main className='main'>
            <AppCrumbs
              items={[
                { href: '/', icon: false, text: 'Dashboard' },
                { href: '/pipelines', icon: false, text: 'Pipelines' },
                { href: '/pipelines/create', icon: false, text: 'Create Pipeline Run', current: true },
              ]}
            />

            <div className='headlineContainer'>
              <Link style={{ display: 'flex', fontSize: '14px', float: 'right', marginLeft: '10px', color: '#777777' }} to='/pipelines'>
                <Icon
                  icon='undo'
                  size={16}
                  style={{ marginRight: '5px', opacity: 0.6 }}
                /> Go Back
              </Link>
              <div style={{ display: 'flex' }}>
                <div>
                  <h1 style={{ margin: 0 }}>
                    Create Pipeline Run
                    <Popover
                      key='popover-help-key-create-pipeline'
                      className='trigger-delete-connection'
                      popoverClassName='popover-help-create-pipeline'
                      position={Position.RIGHT}
                      autoFocus={false}
                      enforceFocus={false}
                      usePortal={false}
                    >
                      <a href='#' rel='noreferrer'><HelpIcon width={19} height={19} style={{ marginLeft: '10px' }} /></a>
                      <>
                        <div style={{ textShadow: 'none', fontSize: '12px', padding: '12px', maxWidth: '300px' }}>
                          <div style={{ marginBottom: '10px', fontWeight: 700, fontSize: '14px' }}>
                            <Icon icon='help' size={16} /> Run Pipeline
                          </div>
                          <p>Need Help? &mdash; Configure the <strong>Data Providers</strong> you want and click
                            <Icon icon='play' size={12} /> <strong>RUN</strong> to trigger a new Pipeline run.
                          </p>
                        </div>
                      </>
                    </Popover>
                  </h1>

                  <p className='page-description mb-0'>Trigger data collection for one or more Data Providers.</p>
                </div>
              </div>
            </div>

            <div className='' style={{ width: '100%', marginTop: '10px', alignSelf: 'flex-start', alignContent: 'flex-start' }}>
              <h2 className='headline'>
                <Icon
                  icon='git-pull'
                  height={16} size={16} color='rgba(0,0,0,0.5)'
                /> Pipeline Name {advancedMode && <>(Advanced)</>}<span className='requiredStar'>*</span>
              </h2>
              <p className='group-caption'>Create a user-friendly name for this Run, or select and use a default auto-generated one.</p>
              <div className='form-group' style={{ maxWidth: '480px', paddingLeft: '22px' }}>
                {isValidPipeline() && (
                  <Icon
                    icon='tick' color={Colors.GREEN5} size={12}
                    style={{ float: 'right', marginTop: '7px', marginLeft: '5px' }}
                  />)}
                {!isValidPipeline() && (
                  <>
                    <Icon
                      icon='exclude-row' color={Colors.RED5}
                      size={12}
                      style={{ float: 'right', marginTop: '7px', marginLeft: '5px' }}
                    />
                  </>
                )}
                <FormGroup
                  disabled={isRunning}
                  label=''
                  labelFor='pipeline-name'
                  className=''
                  contentClassName=''
                  helperText={`RUN DATE = ${today.toLocaleString()}`}
                  fill
                  required
                >

                  <InputGroup
                    id='pipeline-name'
                    disabled={isRunning}
                    placeholder='eg. COLLECTION YYYYMMDDHHMMSS'
                    value={pipelineName}
                    onChange={(e) => setPipelineName(e.target.value)}
                    className={!isValidPipelineForm ? 'input-pipeline-name is-invalid' : 'input-pipeline-name is-valid'}
                    rightElement={
                      <>
                        <Button
                          icon='reset' text='' small
                          minimal
                          onClick={() => resetPipelineName()}
                        />
                        <Popover
                          className='popover-pipeline-menu-trigger'
                          popoverClassName='popover-pipeline-menu'
                          position={Position.RIGHT_BOTTOM}
                          usePortal={true}
                        >
                          <Button text={`${today.toLocaleTimeString()}`} />
                          <>
                            <PipelinePresetsMenu
                              namePrefix={namePrefix}
                              pipelineSuffixes={pipelineSuffixes}
                              setNamePrefix={setNamePrefix}
                              setNameSuffix={setNameSuffix}
                              setRawConfiguration={setRawConfiguration}
                              advancedMode={advancedMode}
                            />
                          </>
                        </Popover>
                        {validationErrors.length > 0 && (
                          <>
                            <div style={{ display: 'block', float: 'right' }}>
                              <Popover
                                key='popover-help-key-validation-errors'
                                className='trigger-validation-errors'
                                popoverClassName='popover-help-validation-errors'
                                position={Position.RIGHT}
                                autoFocus={false}
                                enforceFocus={false}
                                usePortal={false}
                              >
                                <Button
                                  intent={Intent.PRIMARY}
                                  icon={<Icon icon='warning-sign' size={14} color={Colors.ORANGE5} />}
                                  small
                                  style={{ margin: '3px 4px 0 0' }}
                                />
                                <div style={{ padding: '5px', minWidth: '300px', maxWidth: '300px', justifyContent: 'flex-start' }}>
                                  <FormValidationErrors errors={validationErrors} textAlign='left' styles={{ display: 'flex' }} />
                                </div>
                              </Popover>
                            </div>
                          </>
                        )}
                      </>
                    }
                    required
                    // large
                    fill
                  />
                </FormGroup>
              </div>

              {advancedMode && (
                <>
                  <h2 className='headline'>
                    <Icon icon='code' height={16} size={16} color='rgba(0,0,0,0.5)' />{' '}
                    <strong>JSON</strong> Provider Configuration<span className='requiredStar'>*</span>
                  </h2>
                  <p className='group-caption'>Define Plugins and Options manually. Only valid JSON code is allowed.</p>
                  <div style={{ padding: '10px 0', borderBottom: '1px solid rgba(0, 0, 0, 0.08)' }}>
                    <div className='form-group' style={{ paddingLeft: '22px' }}>
                      <Card
                        className='code-editor-card'
                        interactive={false}
                        elevation={Elevation.TWO}
                        style={{ padding: '2px', minWidth: '320px', width: '100%', maxWidth: '640px', marginBottom: '20px' }}
                      >
                        <h3 style={{ borderBottom: '1px solid #eeeeee', margin: 0, padding: '8px 10px' }}>
                          <span style={{ float: 'right', fontSize: '9px', color: '#aaaaaa' }}>application/json</span>
                          TASKS EDITOR
                          {isMultiStagePipeline(runTasksAdvanced) && (
                            <>
                              {' '} &rarr; {' '}
                              <Icon icon='layers' color={Colors.GRAY4} size={14} style={{ marginRight: '5px' }} />
                              <span style={{
                                fontStyle: 'normal',
                                fontWeight: 900,
                                letterSpacing: '1px',
                                color: '#333',
                                fontSize: '11px'
                              }}
                              >
                                MULTI-STAGE <Tag intent={Intent.PRIMARY} style={{ borderRadius: '20px' }}>{runTasksAdvanced.length}</Tag>
                              </span>
                            </>
                          )}
                        </h3>
                        <TextArea
                          growVertically={false}
                          fill={true}
                          className='codeArea'
                          style={{ height: '440px !important', maxWidth: '640px' }}
                          value={rawConfiguration}
                          onChange={(e) => setRawConfiguration(e.target.value)}
                        />
                        {/* @todo: fix bug with @uiw/react-textarea-code-editor in a future release */}
                        {/* <div
                          className='code-editor-wrapper' style={{
                            minHeight: '384px',
                            height: '440px !important',
                            maxWidth: '640px',
                            maxHeight: '384px',
                            overflow: 'hidden',
                            overflowY: 'auto'
                          }}
                        >
                          <CodeEditor
                            value={rawConfiguration}
                            language='json'
                            placeholder='< Please enter JSON configuration with supported Plugins. >'
                            onChange={(e) => setRawConfiguration(e.target.value)}
                            padding={15}
                            minHeight={384}
                            style={{
                              fontSize: 12,
                              backgroundColor: '#f5f5f5',
                              // eslint-disable-next-line max-len
                              fontFamily: 'JetBrains Mono,Source Code Pro,ui-monospace,SFMono-Regular,SF Mono,Consolas,Liberation Mono,Menlo,monospace',
                            }}
                          />
                        </div> */}
                        <div
                          className='code-editor-card-footer'
                          style={{
                            display: 'flex',
                            justifyContent: 'flex-end',
                            padding: '5px',
                            borderTop: '1px solid #eeeeee',
                            fontSize: '11px'
                          }}
                        >
                          <ButtonGroup
                            className='code-editor-controls' style={{
                              borderRadius: '3px',
                              boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.30)'
                            }}
                          >
                            <Popover
                              className='popover-options-menu-trigger'
                              popoverClassName='popover-options-menu'
                              position={Position.TOP}
                              usePortal={true}
                            >
                              <Button
                                disabled={isRunning}
                                icon='cog'
                              />
                              <>
                                <PipelineConfigsMenu
                                  setRawConfiguration={setRawConfiguration}
                                  advancedMode={advancedMode}
                                />
                              </>
                            </Popover>
                            <Button
                              disabled={!isValidConfiguration}
                              small text='Format' icon='align-left'
                              onClick={() => formatRawCode()}
                            />
                            <Button
                              small text='Revert' icon='reset'
                              onClick={() => setRawConfiguration(JSON.stringify([runTasks], null, '  '))}
                            />
                            <Button
                              small text='Clear' icon='eraser'
                              onClick={() => setRawConfiguration('[[]]')}
                            />
                            <Popover
                              className='trigger-code-validation-help'
                              popoverClassName='popover-code-validation-help'
                              position={Position.RIGHT}
                              autoFocus={false}
                              enforceFocus={false}
                              usePortal={false}
                            >
                              <Button
                                intent={isValidConfiguration ? Intent.SUCCESS : Intent.PRIMARY}
                                small
                                text={isValidConfiguration ? 'Valid' : 'Invalid'}
                                icon={isValidConfiguration ? 'confirm' : 'warning-sign'}
                              />
                              <>
                                <div style={{
                                  textShadow: 'none',
                                  fontSize: '12px',
                                  padding: '12px',
                                  minWidth: '300px',
                                  maxWidth: '300px',
                                  maxHeight: '200px',
                                  overflow: 'hidden',
                                  overflowY: 'auto'
                                }}
                                >
                                  {isValidConfiguration
                                    ? (
                                      <>
                                        <Icon
                                          icon='tick' color={Colors.GREEN5} size={16}
                                          style={{ float: 'left', marginRight: '5px' }}
                                        />
                                        <div style={{ fontSize: '13px', fontWeight: 800, marginBottom: '5px' }}>
                                          JSON Configuration Valid
                                        </div>
                                        {isMultiStagePipeline(runTasksAdvanced) && (
                                          <>
                                            <div
                                              className='bp3-elevation-1' style={{
                                                backgroundColor: '#f6f6f6',
                                                padding: '4px 6px',
                                                borderRadius: '3px',
                                                marginBottom: '10px'
                                              }}
                                            >
                                              <Icon icon='layers' color={Colors.GRAY4} size={14} style={{ marginRight: '5px' }} />
                                              <span style={{
                                                fontStyle: 'normal',
                                                fontWeight: 900,
                                                letterSpacing: '1px',
                                                color: '#333',
                                                fontSize: '11px'
                                              }}
                                              >
                                                MULTI-STAGE <Tag>{runTasksAdvanced.length}</Tag>
                                              </span>
                                            </div>
                                            <span style={{ fontSize: '10px' }}>Multi-stage task configuration detected.</span>
                                          </>
                                        )}
                                      </>
                                      )
                                    : (
                                      <>
                                        <Icon
                                          icon='issue' color={Colors.RED5} size={16}
                                          style={{ float: 'left', marginRight: '5px' }}
                                        />
                                        <div style={{ fontSize: '13px', fontWeight: 800, marginBottom: '5px' }}>
                                          Invalid JSON Configuration
                                        </div>
                                        {validationError}
                                      </>
                                      )}
                                </div>
                              </>
                            </Popover>

                          </ButtonGroup>
                        </div>
                      </Card>
                      <div style={{ marginTop: '0', maxWidth: '640px' }}>
                        <div style={{ display: 'flex', minHeight: '34px' }}>
                          <div style={{
                            marginRight: '5px',
                            paddingLeft: '10px',
                            fontWeight: 800,
                            letterSpacing: '2px',
                            color: Colors.GRAY2,
                          }}
                          ><span><Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '2px' }} /> DATA PROVIDERS</span>
                          </div>
                          {detectedProviders.map((provider, pIdx) => (
                            <div className='detected-provider-icon' key={`provider-icon-key-${pIdx}`} style={{ margin: '5px 18px' }}>
                              {ProviderIcons[provider] ? ProviderIcons[provider](20, 20) : <></>}
                            </div>
                          ))}
                          {detectedProviders.length === 0 && (<span style={{ color: Colors.GRAY4 }}>&lt; None Configured &gt;</span>)}
                        </div>
                      </div>
                    </div>
                  </div>
                </>
              )}

              {!advancedMode && (
                <>
                  <h2 className='headline'>
                    <Icon icon='database' height={16} size={16} color='rgba(0,0,0,0.5)' />{' '}
                    Data Providers<span className='requiredStar'>*</span>
                  </h2>
                  <p className='group-caption'>
                    Configure available plugins to enable for this <strong>Pipeline Run</strong>.<br />
                    Turn the switch to the ON position to activate.
                  </p>
                  <div className='data-providers' style={{ marginTop: '8px', width: '100%' }}>
                    {integrations.map((provider) => (
                      <CSSTransition
                        key={`fx-key-provider-${provider.id}`}
                        in={readyProviders.includes(provider.id)}
                        timeout={350}
                        classNames='provider-datarow'
                        unmountOnExit
                      >
                        {/* <div key={`provider-${provider.id}`}> */}
                        <div
                          // eslint-disable-next-line max-len
                          className={`data-provider-row data-provider-${provider.id.toLowerCase()} ${enabledProviders.includes(provider.id) ? 'on' : 'off'}`}
                        >
                          <div className='provider-info'>
                            <div className='provider-icon'>{provider.iconDashboard}</div>
                            <span className='provider-name'>{provider.name}</span>
                            <Tooltip
                              intent={Intent.PRIMARY}
                              content={`Enable ${provider.name}`} position={Position.LEFT} popoverClassName='pipeline-tooltip'
                            >
                              <Switch
                                // alignIndicator={Alignment.CENTER}
                                disabled={isRunning}
                                className={`provider-toggle-switch switch-${provider.id.toLowerCase()}`}
                                innerLabel={!enabledProviders.includes(provider.id) ? 'OFF' : null}
                                innerLabelChecked='ON'
                                checked={enabledProviders.includes(provider.id)}
                                onChange={() => setEnabledProviders(p =>
                                  enabledProviders.includes(provider.id) ? p.filter(p => p !== provider.id) : [...p, provider.id]
                                )}
                              />
                            </Tooltip>
                          </div>
                          <div className='provider-settings'>
                            {/* showProviderSettings(provider.id) */}
                            <ProviderSettings
                              providerId={provider.id}
                              projectId={projectId}
                              owner={owner}
                              repositoryName={repositoryName}
                              connectionId={connectionId}
                              connections={connections}
                              repositories={repositories}
                              selectedConnection={selectedConnection}
                              selectedGithubRepo={selectedGithubRepo}
                              setSelectedConnection={setSelectedConnection}
                              boardId={boardId}
                              gitExtractorUrl={gitExtractorUrl}
                              gitExtractorRepoId={gitExtractorRepoId}
                              refDiffRepoId={refDiffRepoId}
                              refDiffTasks={refDiffTasks}
                              refDiffPairs={refDiffPairs}
                              setProjectId={setProjectId}
                              setOwner={setOwner}
                              setRepositoryName={setRepositoryName}
                              setConnectionId={setConnectionId}
                              setBoardId={setBoardId}
                              setGitExtractorUrl={setGitExtractorUrl}
                              setGitExtractorRepoId={setGitExtractorRepoId}
                              setSelectedGithubRepo={setSelectedGithubRepo}
                              setRefDiffRepoId={setRefDiffRepoId}
                              setRefDiffPairs={setRefDiffPairs}
                              setRefDiffTasks={setRefDiffTasks}
                              isEnabled={isProviderEnabled}
                              isRunning={isRunning}
                            />
                          </div>
                          <div className='provider-actions'>
                            <ButtonGroup minimal rounded='true'>
                              {provider.type === ProviderTypes.INTEGRATION && (
                                <Button className='pipeline-action-btn' minimal onClick={() => history.push(`/integrations/${provider.id}`)}>
                                  <Icon icon='cog' color={Colors.GRAY4} size={16} />
                                </Button>
                              )}
                              <Popover
                                key={`popover-help-key-provider-${provider.id}`}
                                className='trigger-provider-help'
                                popoverClassName='popover-provider-help'
                                position={Position.RIGHT}
                                autoFocus={false}
                                enforceFocus={false}
                                usePortal={true}
                              >
                                <Button className='pipeline-action-btn' minimal><Icon icon='help' color={Colors.GRAY4} size={16} /></Button>
                                <>
                                  <div style={{ textShadow: 'none', fontSize: '12px', padding: '12px', maxWidth: '300px' }}>
                                    <div style={{
                                      marginBottom: '10px',
                                      fontWeight: 700,
                                      fontSize: '14px',
                                    }}
                                    >
                                      <Icon icon='help' size={16} /> {provider.name} Settings
                                    </div>
                                    <p>
                                      Need Help? &mdash; Please enter the required{' '}
                                      <strong>Run Settings</strong> for this data provider.
                                    </p>
                                    {/* specific provider field help notes */}
                                    {(() => {
                                      let helpContext = null
                                      switch (provider.id) {
                                        case Providers.GITLAB:
                                          helpContext = (
                                            <img
                                              src={GitlabHelpNote}
                                              alt={provider.name} style={{ maxHeight: '64px', maxWidth: '100%' }}
                                            />
                                          )
                                          break
                                        case Providers.JENKINS:
                                          helpContext = <strong>(Options not required)</strong>
                                          break
                                        case Providers.JIRA:
                                          helpContext = (
                                            <img
                                              src={JiraHelpNote}
                                              alt={provider.name} style={{ maxHeight: '64px', maxWidth: '100%' }}
                                            />
                                          )
                                          break
                                        case Providers.GITHUB:
                                          helpContext = (
                                            <img
                                              src={GithubHelpNote}
                                              alt={provider.name} style={{ maxHeight: '64px', maxWidth: '100%' }}
                                            />
                                          )
                                          break
                                        case Providers.GITEXTRACTOR:
                                          helpContext = (
                                            <>
                                              <div><strong>GitExtractor README</strong></div>
                                              <p>This plugin extract commits and references from a remote or local git repository.</p>
                                              <a
                                                className='bp3-button bp3-small'
                                                rel='noreferrer'
                                                target='_blank'
                                                href='https://github.com/apache/incubator-devlake/tree/main/plugins/gitextractor'
                                              >Learn More
                                              </a>
                                            </>
                                          )
                                          break
                                        case Providers.REFDIFF:
                                          helpContext = (
                                            <>
                                              <div><strong>RefDiff README</strong></div>
                                              <p>You need to run gitextractor before the refdiff plugin.</p>
                                              <a
                                                className='bp3-button bp3-small'
                                                rel='noreferrer'
                                                target='_blank'
                                                href='https://github.com/apache/incubator-devlake/tree/main/plugins/refdiff'
                                              >Learn More
                                              </a>
                                            </>
                                          )
                                          break
                                      }
                                      return helpContext
                                    })()}
                                  </div>
                                </>
                              </Popover>
                            </ButtonGroup>
                          </div>
                        </div>
                        {/* </div> */}
                      </CSSTransition>
                    ))}
                  </div>
                </>
              )}

            </div>
            <div className='blueprint-options' style={{ alignSelf: 'flex-start', justifyContent: 'flex-start', marginBottom: '30px' }}>
              <div style={{ marginTop: '10px', display: 'flex', justifyContent: 'flex-start', alignItems: 'center', alignContent: 'center' }}>
                <h2 className='headline' style={{ color: enableAutomation ? Colors.BLACK : Colors.GRAY2 }}>
                  <Icon
                    icon='calendar'
                    height={16} size={16} color='rgba(0,0,0,0.5)'
                  /> Automate Pipeline

                </h2>
                <Switch
                  disabled={advancedMode ? !isValidAdvancedPipeline() : !isValidPipeline()}
                  style={{ display: 'flex', alignSelf: 'center', margin: '13px 0 0 15px' }}
                  checked={enableAutomation}
                  onChange={(e) => setEnableAutomation(a => !a)} label={false}
                />
              </div>
              <p className='group-caption'>
                Automatically run this pipeline configuration by setting up a recurring <strong>Blueprint</strong>.
              </p>
              {!saveBlueprintComplete && (
                <Button
                  ref={addBlueprintRef}
                  disabled={!enableAutomation || (advancedMode ? !isValidAdvancedPipeline() : !isValidPipeline())}
                  intent={enableAutomation ? Intent.WARNING : Intent.NONE}
                  small text='Add Blueprint'
                  icon='plus'
                  style={{ marginLeft: '25px' }}
                  onClick={() => setBlueprintDialogIsOpen(opened => !opened)}
                />
              )}
              {saveBlueprintComplete && (
                <ButtonGroup>
                  <Button
                    ref={addBlueprintRef}
                    disabled={!enableAutomation}
                    intent={enableAutomation ? Intent.WARNING : Intent.NONE}
                    small
                    text={saveBlueprintComplete.name}
                    icon='bold'
                    style={{ marginLeft: '25px' }}
                    onClick={() => setBlueprintDialogIsOpen(opened => !opened)}
                  />
                  <Button
                    disabled={isDeletingBlueprint}
                    icon='trash'
                    text='Delete'
                    small
                    onClick={() => deleteBlueprint(saveBlueprintComplete)}
                  />
                </ButtonGroup>
              )}
            </div>
            <div style={{ display: 'flex', marginTop: '32px', width: '100%', justifyContent: 'flex-start' }}>
              {validationErrors.length > 0 && (
                <FormValidationErrors errors={validationErrors} />
              )}
            </div>
            <div style={{ display: 'flex', width: '100%', justifyContent: 'flex-start', alignItems: 'flex-start' }}>
              <Button
                id='btn-run-pipeline'
                className='btn-pipeline btn-run-pipeline' icon='play' intent='primary'
                disabled={advancedMode ? !isValidAdvancedPipeline() : !isValidPipeline()}
                onClick={runPipeline}
                loading={isRunning}
              ><strong>Run</strong> Pipeline
              </Button>
              <Tooltip content='Manage Pipelines' position={Position.TOP}>
                <Button
                  onClick={() => history.push('/pipelines')}
                  className='btn-pipeline btn-view-jobs'
                  icon='pulse' minimal style={{ marginLeft: '5px' }}
                >View All Pipelines
                </Button>
              </Tooltip>
              <Button
                className='btn-pipeline btn-reset-pipeline'
                icon='eraser' minimal style={{ marginLeft: '5px' }}
                onClick={resetConfiguration}
              >Reset
              </Button>
              <div style={{ padding: '7px 5px 0 50px' }}>
                <Tooltip content='Advanced Pipeline Mode' position={Position.TOP}>
                  <Switch
                    className='advanced-mode-toggleswitch'
                    intent={Intent.DANGER}
                    checked={advancedMode}
                    onChange={() => setAdvancedMode(t => !t)}
                    labelElement={
                      <>
                        <span style={{
                          fontSize: '14px',
                          fontWeight: 800,
                          display: 'inline-block',
                          whiteSpace: 'nowrap'
                        }}
                        >Advanced Mode
                        </span><br />
                        <strong style={{ color: !advancedMode ? Colors.GRAY3 : '' }}>Raw JSON Trigger</strong>
                      </>
                    }
                  />
                </Tooltip>
              </div>
            </div>
            <p style={{ margin: '5px 3px', alignSelf: 'flex-start', fontSize: '10px' }}>
              Visit the <a href='#'><strong>All Jobs</strong></a> section to monitor complete pipeline activity.<br />
              Once you run this pipeline, you’ll be redirected to collection status.
            </p>
            {advancedMode && (
              <div style={{ alignSelf: 'flex-start' }}>
                <h4 style={{
                  marginBottom: '8px',
                  fontSize: '12px',
                  fontWeight: 700,
                }}
                ><Icon icon='issue' size={12} style={{ marginBottom: '2px' }} /> <span>Expert Use Only</span>
                </h4>
                <p style={{ fontSize: '10px' }}>
                  Trigger a manual Pipeline with <a href='#'><strong>JSON Configuration</strong></a>.<br />
                  Please review the{' '}
                  <a
                    href='https://github.com/apache/incubator-devlake/wiki/How-to-use-the-triggers-page' target='_blank'
                    rel='noreferrer'
                    style={{
                      fontWeight:
                        'bold',
                      color: '#E8471C',
                      textDecoration: 'underline'
                    }}
                  >
                    Documentation
                  </a> on creating complex Pipelines.
                </p>
              </div>
            )}
          </main>
        </Content>
      </div>
      <PipelineIndicator
        pipeline={pipelineRun}
        graphsUrl={GRAFANA_URL}
        onFetch={fetchPipeline}
        onCancel={cancelPipeline}
        onView={() => history.push(`/pipelines/activity/${pipelineRun.ID}`)}
      />
      <AddBlueprintDialog
        isLoading={isFetchingAllPipelines}
        isOpen={blueprintDialogIsOpen}
        setIsOpen={setBlueprintDialogIsOpen}
        name={name}
        cronConfig={cronConfig}
        customCronConfig={customCronConfig}
        getNextRunDate={getNextRunDate}
        enable={enable}
        tasks={blueprintTasks}
        draftBlueprint={draftBlueprint}
        setDraftBlueprint={setDraftBlueprint}
        setBlueprintName={setBlueprintName}
        setCronConfig={setCronConfig}
        setCustomCronConfig={setCustomCronConfig}
        setEnableBlueprint={setEnableBlueprint}
        setBlueprintTasks={setBlueprintTasks}
        createCron={createCron}
        saveBlueprint={saveBlueprint}
        isSaving={isSaving}
        isValidBlueprint={isValidBlueprint}
        fieldHasError={fieldHasError}
        getFieldError={getFieldError}
        pipelines={pipelineTemplates}
        selectedPipelineTemplate={selectedPipelineTemplate}
        setSelectedPipelineTemplate={setSelectedPipelineTemplate}
        detectedProviders={detectedProviderTasks}
        getCronPreset={getCronPreset}
        getCronPresetByConfig={getCronPresetByConfig}
        tasksLocked={true}
      />

    </>
  )
}

export default CreatePipeline
