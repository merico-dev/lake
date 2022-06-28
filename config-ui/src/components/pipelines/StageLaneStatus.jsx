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
import React from 'react'
import { Colors, Icon, Intent, ProgressBar } from '@blueprintjs/core'

const StageLaneStatus = (props) => {
  const {
    stage,
    sK = 1,
    duration = '0 mins',
    isStageActive = () => {},
    isStagePending = () => {},
    isStageCompleted = () => {},
    isStageFailed = () => {},
    calculateStageLaneProgress = () => {},
    getRunningTaskCount = () => {},
    getTotalTasksCount = () => {},
    getCompletedTaskCount = () => {}
  } = props

  return (
    <>
      <div
        className='stage-footer' style={{
          padding: '5px 10px',
          marginBottom: '4px',
          alignSelf: 'flex-end',
          marginTop: 'auto',
          fontSize: '10px',
          color: '#999999',
          alignContent: 'flex-end',
          textAlign: 'right'
        }}
      >
        <div
          className='stage-status'
          style={{
            fontWeight: 900,
            fontSize: '10px',
            letterSpacing: '1px',
            color: isStageCompleted(sK) ? Colors.GREEN5 : Colors.GRAY5
          }}
        >
          {isStageActive(sK) && <span style={{ color: Colors.BLACK }}> ACTIVE</span>}
          {isStageCompleted(sK) && <span style={{ color: Colors.GREEN5 }}>COMPLETED</span>}
          {isStageFailed(sK) && <span style={{ color: Colors.RED5 }}>FAILED</span>}
          {isStagePending(sK) && <>WAITING</>}
          <span style={{ fontWeight: 500, color: isStageActive(sK) ? '#000' : 'inherit', opacity: 0.6 }}>
            {' '}&middot;{' '}
            {isStageCompleted(sK) ? getCompletedTaskCount(stage) : getRunningTaskCount(stage)}/{getTotalTasksCount(stage)}
          </span>
        </div>
        <div className='stage-caption'>
          {isStageActive(sK) && <>Stage Running ~{duration}</>}
          {/* {isStageFailed(sK) && <>Stage Failed</>} */}
          {(isStageCompleted(sK) || isStageFailed(sK)) && <>{duration.startsWith('0') ? '< 1min' : duration}</>}
          {isStagePending(sK) && <><Icon icon='more' color={Colors.GRAY5} size={12} /></>}
        </div>
      </div>
      {isStageActive(sK) && (
        <ProgressBar
          className='stage-lane-progressbar'
          stripes={true}
          intent={Intent.SUCCESS} value={calculateStageLaneProgress(stage)} style={{ borderRadius: 0 }}
        />
      )}
    </>
  )
}

export default StageLaneStatus
