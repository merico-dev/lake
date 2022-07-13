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
import { useCallback, useEffect, useState } from 'react'
import { Providers, } from '@/data/Providers'
import { BlueprintMode } from '@/data/NullBlueprint'

function usePipelineValidation ({
  activeStep,
  enabledProviders = [],
  pipelineName,
  projectId,
  projects = [],
  boardId,
  boards = [],
  owner,
  repositoryName,
  gitExtractorUrl,
  gitExtractorRepoId,
  refDiffRepoId,
  refDiffTasks = [],
  refDiffPairs = [],
  connectionId,
  tasks = [],
  tasksAdvanced = [],
  advancedMode,
  mode = null,
  connection,
  entities = []
}) {
  const [errors, setErrors] = useState([])
  const [isValid, setIsValid] = useState(false)
  const [detectedProviders, setDetectedProviders] = useState([])
  const [allowedProviders, setAllowedProviders] = useState([
    Providers.JIRA,
    Providers.GITLAB,
    Providers.JENKINS,
    Providers.GITHUB,
    Providers.REFDIFF,
    Providers.GITEXTRACTOR,
    Providers.FEISHU,
    Providers.AE,
    Providers.DBT,
    Providers.STARROCKS
  ])

  const clear = () => {
    setErrors([])
  }

  const validateNumericSet = (set = []) => {
    return Array.isArray(set) ? set.every(i => !isNaN(i)) : false
  }

  const validateRepositoryName = (set = []) => {
    const repoRegExp = /([a-z0-9_-]){2,}\/([a-z0-9_-]){2,}/gi
    return set.every(i => i.match(repoRegExp))
  }

  const validate = useCallback(() => {
    const errs = []
    console.log('>> VALIDATING PIPELINE RUN ', pipelineName)

    // if (!pipelineName || pipelineName.length <= 2) {
    //   errs.push('Name: Enter a valid Pipeline Name')
    // }

    if (mode === BlueprintMode.NORMAL && enabledProviders.includes(Providers.GITLAB) && projects[connection?.id]?.length === 0) {
      errs.push('GitLab: Enter one valid Project ID (Numeric)')
    }

    if (mode === BlueprintMode.NORMAL && enabledProviders.includes(Providers.GITLAB) && projects[connection?.id]?.length === 0) {
      errs.push('GitLab: Enter one valid Project ID (Numeric)')
    }

    if (mode === BlueprintMode.NORMAL && enabledProviders.includes(Providers.GITLAB) && !validateNumericSet(projects[connection?.id])) {
      errs.push('GitLab: Entered Project ID is NOT numeric!')
    }

    if (mode === BlueprintMode.NORMAL && enabledProviders.includes(Providers.JIRA) && boards[connection?.id]?.length === 0) {
      errs.push('JIRA: Enter one valid Board ID (Numeric)')
    }

    if (mode === BlueprintMode.NORMAL && enabledProviders.includes(Providers.JIRA) && !validateNumericSet(boards[connection?.id])) {
      errs.push('JIRA: Entered Board ID is NOT numeric!')
    }

    if (mode === BlueprintMode.NORMAL && enabledProviders.includes(Providers.GITHUB) && projects[connection?.id]?.length === 0) {
      errs.push('GitHub: Enter one valid Project (Repository)')
    }

    if (mode === BlueprintMode.NORMAL && enabledProviders.includes(Providers.GITHUB) && !validateRepositoryName(projects[connection?.id])) {
      errs.push('GitHub: Entered Repository Name Invalid Format (owner/repo)')
    }

    // @todo: fix this, technically this rule isn't valid connectionId is only for 1 provider
    // if (mode === null && enabledProviders.includes(Providers.GITLAB) && (!connectionId || isNaN(connectionId))) {
    //   errs.push('GitLab: Select a valid Connection ID (Numeric)')
    // }

    if (mode === null && enabledProviders.includes(Providers.GITLAB) && (!projectId || projectId.length !== 1 || projectId.toString() === '')) {
      errs.push('GitLab: Enter one valid Project ID (Numeric)')
    }

    if (mode === null && enabledProviders.includes(Providers.GITLAB) && !validateNumericSet(projectId)) {
      errs.push('GitLab: Entered Project ID is NOT numeric!')
    }

    if (mode === null && enabledProviders.includes(Providers.JIRA) && (!connectionId || isNaN(connectionId))) {
      errs.push('JIRA: Select a valid Connection ID (Numeric)')
    }

    if (mode === null && enabledProviders.includes(Providers.JIRA) && (!boardId || boardId.length !== 1 || boardId.toString() === '')) {
      errs.push('JIRA: Enter one valid Board ID (Numeric)')
    }

    if (mode === null && enabledProviders.includes(Providers.JIRA) && !validateNumericSet(boardId)) {
      errs.push('JIRA: Entered Board ID is NOT numeric!')
    }

    if (mode === null && enabledProviders.includes(Providers.GITHUB) && (!owner || owner <= 2)) {
      errs.push('GitHub: Owner/Developer is required')
    }

    if (mode === null && enabledProviders.includes(Providers.GITHUB) && (owner.match(/^[a-zA-Z0-9_-]+$/g) === null)) {
      errs.push('GitHub: Owner invalid format')
    }

    if (mode === null && enabledProviders.includes(Providers.GITHUB) && !repositoryName) {
      errs.push('GitHub: Repository Name is required')
    }

    if (mode === null && enabledProviders.includes(Providers.GITHUB) && repositoryName.match(/^[a-zA-Z0-9._-]+$/g) === null) {
      errs.push('GitHub: Repository name invalid format')
    }

    if (mode === null && enabledProviders.includes(Providers.GITEXTRACTOR) && !gitExtractorUrl) {
      errs.push('GitExtractor: Repository Git URL is required')
    }

    if (mode === null && enabledProviders.includes(Providers.GITEXTRACTOR) &&
        gitExtractorUrl.toLowerCase().match(/^(http:\/\/|https:\/\/|ssh:\/\/|git@)+/g) === null) {
      errs.push('GitExtractor: Repository Git URL must be valid HTTP/S, SSH or Git@ protocol')
    }

    if (mode === null && enabledProviders.includes(Providers.GITEXTRACTOR) && !gitExtractorUrl.toLowerCase().endsWith('.git')) {
      errs.push('GitExtractor: Invalid Git URL Extension')
    }

    if (mode === null && enabledProviders.includes(Providers.GITEXTRACTOR) && !gitExtractorRepoId) {
      errs.push('GitExtractor: Repository Column ID Code is required')
    }

    if (mode === null && enabledProviders.includes(Providers.REFDIFF) && !refDiffRepoId) {
      errs.push('RefDiff: Repository Column ID Code is required')
    }

    if (mode === null && enabledProviders.includes(Providers.REFDIFF) && refDiffTasks.length === 0) {
      errs.push('RefDiff: Please select at least ONE (1) Plugin Task')
    }

    if (mode === null && enabledProviders.includes(Providers.REFDIFF) && refDiffPairs.length === 0) {
      errs.push('RefDiff: Please enter at least ONE (1) Tag Ref Pair')
    }

    if (enabledProviders.length === 0) {
      errs.push(mode === null ? 'Pipeline: Invalid/Empty Configuration' : 'No Connections Configured!')
    }
    setErrors(errs)
  }, [
    enabledProviders,
    pipelineName,
    projectId,
    boardId,
    owner,
    repositoryName,
    gitExtractorUrl,
    gitExtractorRepoId,
    refDiffRepoId,
    refDiffTasks,
    refDiffPairs,
    connectionId,
    boards,
    projects
  ])

  const validateAdvanced = useCallback(() => {
    const errs = []
    let parsed = []
    if (advancedMode) {
      console.log('>> VALIDATING ADVANCED PIPELINE RUN ', tasksAdvanced, pipelineName)

      if (Array.isArray(tasksAdvanced)) {
        // eslint-disable-next-line max-len
        setDetectedProviders([...new Set(tasksAdvanced?.flat().filter(aT => allowedProviders.includes(aT.Plugin || aT.plugin)).map(p => p.Plugin || p.plugin))])
      }

      if (!pipelineName || pipelineName.length <= 2) {
        errs.push('Name: Enter a valid Pipeline Name')
      }

      try {
        // eslint-disable-next-line no-unused-vars
        parsed = JSON.parse(JSON.stringify(tasksAdvanced))
      } catch (e) {
        errs.push('Advanced Pipeline: Invalid JSON Configuration')
      }

      if (Array.isArray(tasksAdvanced) && tasksAdvanced?.flat().length === 0) {
        errs.push('Advanced Pipeline: Invalid/Empty Configuration')
      }

      if (!Array.isArray(tasksAdvanced) || !Array.isArray(tasksAdvanced[0])) {
        errs.push('Advanced Pipeline: Invalid Tasks Array Structure!')
      }

      if (Array.isArray(tasksAdvanced) && !tasksAdvanced?.flat().every(aT => allowedProviders.includes(aT.Plugin || aT.plugin))) {
        errs.push('Advanced Pipeline: Unsupported Data Provider Plugin Detected!')
      }

      console.log('>>> Advanced Pipeline Validation Errors? ...', errs)
    }
    setErrors(errs)
  }, [
    advancedMode,
    tasksAdvanced,
    pipelineName,
    allowedProviders
  ])

  useEffect(() => {
    console.log('>>> PIPELINE RUN FORM ERRORS...', errors)
    setIsValid(errors.length === 0)
    if (errors.length > 0) {
      // ToastNotification.clear()
    }
  }, [errors])

  useEffect(() => {
    console.log('>>> DETECTED PLUGIN PROVIDERS...', detectedProviders)
  }, [detectedProviders])

  return {
    errors,
    setErrors,
    isValid,
    validate,
    validateAdvanced,
    clear,
    setAllowedProviders,
    allowedProviders,
    detectedProviders
  }
}

export default usePipelineValidation
