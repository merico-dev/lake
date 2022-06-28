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
import { CSSTransition } from 'react-transition-group'
import { useHistory, useLocation, Link } from 'react-router-dom'
import dayjs from '@/utils/time'

const BlueprintDetail = (props) => {
  const history = useHistory()
  const { bId } = useParams()

  const [blueprintId, setBlueprintId] = useState()

  const {
    // eslint-disable-next-line no-unused-vars
    blueprint,
    blueprints,
    name,
    cronConfig,
    customCronConfig,
    cronPresets,
    tasks,
    detectedProviderTasks,
    enable,
    setName: setBlueprintName,
    setCronConfig,
    setCustomCronConfig,
    setTasks: setBlueprintTasks,
    setDetectedProviderTasks,
    setEnable: setEnableBlueprint,
    isFetching: isFetchingBlueprints,
    isSaving,
    isDeleting,
    createCronExpression: createCron,
    getCronSchedule: getSchedule,
    getCronPreset,
    getCronPresetByConfig,
    getNextRunDate,
    activateBlueprint,
    deactivateBlueprint,
    // eslint-disable-next-line no-unused-vars
    fetchBlueprint,
    fetchAllBlueprints,
    saveBlueprint,
    deleteBlueprint,
    saveComplete,
    deleteComplete
  } = useBlueprintManager()


  useEffect(() => {
    setBlueprintId(bId)
    console.log('>>> REQUESTED BLUEPRINT ID ===', bId)
  }, [bId])

  useEffect(() => {
    if (blueprintId) {
      // @todo: enable blueprint data fetch
      // fetchBlueprint(blueprintId)
    }

  }, [blueprintId, fetchBlueprint])

  return (
    <>
      <div className="container">
        <Nav />
        <Sidebar />
        <Content>
          <main className="main">

          </main>
        </Content>
      </div>
    </>
  )
}

export default BlueprintDetail
