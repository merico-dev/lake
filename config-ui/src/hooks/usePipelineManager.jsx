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
import { DEVLAKE_ENDPOINT } from '@/utils/config'
import request from '@/utils/request'
import { NullPipelineRun } from '@/data/NullPipelineRun'
import { ToastNotification } from '@/components/Toast'
import { Providers } from '@/data/Providers'
// import { integrationsData } from '@/data/integrations'

function usePipelineManager (myPipelineName = `COLLECTION ${Date.now()}`, initialTasks = []) {
  // const [integrations, setIntegrations] = useState(integrationsData)
  const [pipelineName, setPipelineName] = useState(myPipelineName ?? `COLLECTION ${Date.now()}`)
  const [isFetching, setIsFetching] = useState(false)
  const [isFetchingAll, setIsFetchingAll] = useState(false)
  const [isRunning, setIsRunning] = useState(false)
  const [isCancelling, setIsCancelling] = useState(false)
  const [errors, setErrors] = useState([])
  const [settings, setSettings] = useState({
    name: pipelineName,
    plan: [
      [...initialTasks]
    ]
  })

  const [pipelines, setPipelines] = useState([])
  const [pipelineCount, setPipelineCount] = useState(0)
  const [activePipeline, setActivePipeline] = useState(NullPipelineRun)
  const [lastRunId, setLastRunId] = useState(null)
  const [pipelineRun, setPipelineRun] = useState(NullPipelineRun)
  const [allowedProviders, setAllowedProviders] = useState([
    Providers.JIRA,
    Providers.GITLAB,
    Providers.JENKINS,
    Providers.GITHUB,
    Providers.REFDIFF,
    Providers.GITEXTRACTOR,
    Providers.FEISHU,
    Providers.AE,
    Providers.DBT
  ])

  const runPipeline = useCallback(() => {
    console.log('>> RUNNING PIPELINE....')
    try {
      setIsRunning(true)
      setErrors([])
      ToastNotification.clear()
      console.log('>> DISPATCHING PIPELINE REQUEST', settings)
      const run = async () => {
        const p = await request.post(`${DEVLAKE_ENDPOINT}/pipelines`, settings)
        const t = await request.get(`${DEVLAKE_ENDPOINT}/pipelines/${p.data?.ID || p.data?.id}/tasks`)
        console.log('>> RAW PIPELINE DATA FROM API...', p.data)
        setPipelineRun({ ...p.data, ID: p.data?.ID || p.data?.id, tasks: [...t.data.tasks] })
        setLastRunId(p.data?.ID || p.data?.id)
        ToastNotification.show({ message: `Created New Pipeline - ${pipelineName}.`, intent: 'danger', icon: 'small-tick' })
        setTimeout(() => {
          setIsRunning(false)
        }, 500)
      }
      run()
    } catch (e) {
      setIsRunning(false)
      setErrors([e.message])
      console.log('>> FAILED TO RUN PIPELINE!!', e)
    }
  }, [pipelineName, settings])

  const cancelPipeline = useCallback((pipelineID) => {
    try {
      setIsCancelling(true)
      setErrors([])
      ToastNotification.clear()
      console.log('>> DISPATCHING CANCEL PIPELINE REQUEST, RUN ID =', pipelineID)
      const cancel = async () => {
        const c = await request.delete(`${DEVLAKE_ENDPOINT}/pipelines/${pipelineID}`)
        console.log('>> RAW PIPELINE CANCEL RUN RESPONSE FROM API...', c)
        setPipelineRun(NullPipelineRun)
        ToastNotification.show({ message: `Pipeline RUN ID - ${pipelineID} Cancelled`, intent: 'danger', icon: 'small-tick' })
        setTimeout(() => {
          setIsCancelling(false)
        }, 500)
      }
      cancel()
    } catch (e) {
      setIsCancelling(false)
      setErrors([e.message])
      console.log('>> FAILED TO FETCH CANCEL PIPELINE RUN!!', pipelineID, e)
    }
  }, [])

  const fetchPipeline = useCallback((pipelineID, refresh = false) => {
    if (!pipelineID) {
      console.log('>> !ABORT! Pipeline ID Missing! Aborting Fetch...')
      // return ToastNotification.show({ message: 'Pipeline ID Missing! Aborting Fetch...', intent: 'danger', icon: 'warning-sign' })
    }
    try {
      setIsFetching(true)
      setErrors([])
      ToastNotification.clear()
      console.log('>> FETCHING PIPELINE RUN DETAILS...')
      const fetch = async () => {
        const p = await request.get(`${DEVLAKE_ENDPOINT}/pipelines/${pipelineID}`)
        const t = await request.get(`${DEVLAKE_ENDPOINT}/pipelines/${pipelineID}/tasks`)
        console.log('>> RAW PIPELINE RUN DATA FROM API...', p.data)
        console.log('>> RAW PIPELINE TASKS DATA FROM API...', t.data)
        setActivePipeline({
          ...p.data,
          ID: p.data.ID || p.data.id,
          tasks: [...t.data.tasks]
        })
        setPipelineRun((pR) => refresh ? { ...p.data, ID: p.data.id, tasks: [...t.data.tasks] } : pR)
        setLastRunId((lrId) => refresh ? p.data?.ID : lrId)
        // ToastNotification.show({ message: `Fetched Pipeline ID - ${p.data?.ID}.`, intent: 'danger', icon: 'small-tick' })
        setTimeout(() => {
          setIsFetching(false)
        }, 500)
      }
      fetch()
    } catch (e) {
      setIsFetching(false)
      setErrors([e.message])
      setActivePipeline(NullPipelineRun)
      console.log('>> FAILED TO FETCH PIPELINE RUN!!', e)
    }
  }, [])

  const fetchPipelineTasks = useCallback(() => {
    try {
      setIsFetching(true)
      setErrors([])
      ToastNotification.clear()
    } catch (e) {
      setIsFetching(false)
      setErrors([e.message])
      console.log('>> FAILED TO FETCH PIPELINE RUN TASKS!!', e)
    }
  }, [])

  const fetchAllPipelines = useCallback((status = null, fetchTimeout = 500) => {
    try {
      setIsFetchingAll(true)
      setErrors([])
      ToastNotification.clear()
      console.log('>> FETCHING ALL PIPELINE RUNS...')
      const fetchAll = async () => {
        let queryParams = '?'
        queryParams += status && ['TASK_COMPLETED', 'TASK_RUNNING', 'TASK_FAILED'].includes(status)
          ? `status=${status}&`
          : ''
        const p = await request.get(`${DEVLAKE_ENDPOINT}/pipelines${queryParams}`)
        console.log('>> RAW PIPELINES RUN DATA FROM API...', p.data?.pipelines)
        let pipelines = p.data && p.data.pipelines ? [...p.data.pipelines] : []
        pipelines = pipelines.map(p => ({ ...p, ID: p.ID || p.id }))
        setPipelines(pipelines)
        setPipelineCount(p.data ? p.data.count : 0)
        // ToastNotification.show({ message: `Fetched All Pipelines`, intent: 'danger', icon: 'small-tick' })
        setTimeout(() => {
          setIsFetchingAll(false)
        }, fetchTimeout)
      }
      fetchAll()
    } catch (e) {
      setIsFetchingAll(false)
      setErrors([e.message])
      setPipelines([])
      setPipelineCount(0)
      console.log('>> FAILED TO FETCH ALL PIPELINE RUNS!!', e)
    }
  }, [])

  const buildPipelineStages = useCallback((tasks = [], outputArray = false) => {
    let stages = {}
    let stagesArray = []
    tasks?.forEach(tS => {
      stages = {
        ...stages,
        [tS.pipelineRow]: tasks?.filter(t => t.pipelineRow === tS.pipelineRow)
      }
    })
    stagesArray = Object.values(stages)
    console.log('>>> BUILDING PIPELINE STAGES...', tasks, stages, stagesArray)
    return outputArray ? stagesArray : stages
  }, [])

  const detectPipelineProviders = useCallback((tasks, providers = allowedProviders) => {
    return [...tasks?.flat().filter(aT => providers.includes(aT.Plugin || aT.plugin)).map(p => p.Plugin || p.plugin)]
  }, [allowedProviders])

  useEffect(() => {
    console.log('>> PIPELINE MANAGER - RECEIVED RUN/TASK SETTINGS', settings)
  }, [settings])

  useEffect(() => {

  }, [pipelineName, initialTasks])

  return {
    errors,
    isRunning,
    isFetching,
    isFetchingAll,
    isCancelling,
    pipelineName,
    settings,
    setSettings,
    setPipelineName,
    pipelineRun,
    activePipeline,
    pipelines,
    pipelineCount,
    lastRunId,
    runPipeline,
    cancelPipeline,
    fetchPipeline,
    fetchPipelineTasks,
    fetchAllPipelines,
    buildPipelineStages,
    detectPipelineProviders,
    allowedProviders,
    setAllowedProviders
  }
}

export default usePipelineManager
