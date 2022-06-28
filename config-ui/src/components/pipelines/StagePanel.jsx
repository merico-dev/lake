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
// import { CSSTransition } from 'react-transition-group'
import {
  Card,
  Button, Icon,
  ButtonGroup,
  Elevation,
  Colors,
  Spinner,
  Intent
  // Alignment, Classes, Spinner
} from '@blueprintjs/core'

const StagePanel = (props) => {
  const {
    activePipeline,
    // pipelineReady = false,
    stages, activeStageId = 1, isLoading = false
  } = props

  const getActiveStageDisplayColor = (status) => {
    let color = 'rgb(0, 102, 255)'
    switch (status) {
      case 'TASK_COMPLETED':
        color = Colors.GREEN4
        break
      case 'TASK_FAILED' :
        color = Colors.RED4
        break
      case 'TASK_CREATED':
        color = Colors.GRAY4
        break
      case 'TASK_RUNNING':
      default:
        color = 'rgb(0, 102, 255)'
        break
    }
    return color
  }

  return (
    <>
      {/* <CSSTransition
        in={pipelineReady}
        timeout={350}
        classNames='activity-panel'
      > */}
      <Card
        elevation={isLoading ? Elevation.THREE : Elevation.TWO}
        className='stage-panel-card'
        style={{
          transition: 'all 0.3s ease-out',
          display: 'flex',
          width: '100%',
          minWidth: '630px',
          justifySelf: 'flex-start',
          marginBottom: '8px',
          padding: 0,
          backgroundColor: activePipeline.status === 'TASK_COMPLETED' ? 'rgba(245, 255, 250, 0.99)' : 'inherit',
          overflow: 'hidden',
          whiteSpace: 'nowrap',
          textOverflow: 'ellipsis'
        }}
      >

        <ButtonGroup style={{ backgroundColor: 'transparent', zIndex: '-1' }}>
          <Button minimal active style={{ width: '32px', backgroundColor: '#eeeeee', padding: 0 }}>
            <div style={{
              margin: 0,
              display: 'flex',
              position: 'relative',
              justifyContent: 'center',
              alignItems: 'center',
              alignContent: 'center'
            }}
            >
              {isLoading && (
                <span style={{
                  position: 'absolute',
                  width: '24px',
                  height: '24px',
                  marginLeft: '4px',
                  display: 'flex',
                  justifyContent: 'center'
                }}
                >
                  <Spinner size={20} intent={Intent.PRIMARY} />
                </span>)}
              {(() => {
                let statusIcon = null
                switch (activePipeline.status) {
                  case 'TASK_COMPLETED':
                    statusIcon = (
                      <Icon
                        icon='tick-circle'
                        size={24}
                      />
                    )
                    break
                  case 'TASK_FAILED':
                    statusIcon = (
                      <Icon
                        icon='error'
                        size={24}
                      />
                    )
                    break
                  case 'TASK_RUNNING':
                  default:
                    statusIcon = (
                      <Icon
                        style={{ margin: 0, padding: 0, float: 'left' }}
                        icon='refresh'
                        size={24}
                      />
                    )
                    break
                }
                return !isLoading && (<span style={{ position: 'absolute', marginLeft: '3px', width: '24px', height: '24px' }}>{statusIcon}</span>)
              })()}
            </div>
          </Button>
          <Button
            minimal
            style={{
              position: 'relative',
              backgroundColor: '#eeeeee',
              paddingRight: '20px',
            }}
          >
            <h3
              className='stage-panel-stage-name'
              style={{ margin: 0, color: Colors.BLACK }}
            >
              Active Stage
            </h3>
          </Button>
          <Button
            className='stage-panel-stage-endcap'
            minimal
            style={{
              marginLeft: '1px',
              background: '#ffffff!!important',
              width: 0,
              height: 0,
              borderTop: '16px solid transparent',
              borderBottom: '16px solid transparent',
              borderLeft: '16px solid #eeeeee',
              pointerEvents: 'none'
            }}
          />
          <h3
            className='active-stage-panel-display'
            style={{
              color: getActiveStageDisplayColor(activePipeline.status),
              textTransform: 'uppercase',
              lineHeight: '33px',
              margin: 0,
              fontWeight: 800,
              fontSize: '13px',
              letterSpacing: '2px',
              justifySelf: 'flex-start'
            }}
          >
            Stage {activeStageId}
          </h3>
          {Object.keys(stages).length === 0 && (
            <>
              <Button
                disabled
                minimal
                style={{
                  position: 'relative',
                  backgroundColor: '#eeeeee',
                  paddingRight: '50px',
                }}
              >
                <h3
                  className='stage-panel-stage-name'
                  style={{ margin: 0, fontSize: '18px', color: Colors.GRAY3 }}
                >
                  No Stages
                </h3>
              </Button>
              <Button
                className='stage-panel-stage-endcap'
                minimal
                style={{
                  marginLeft: '1px',
                  background: '#ffffff!!important',
                  width: 0,
                  height: 0,
                  borderTop: '16px solid transparent',
                  borderBottom: '16px solid transparent',
                  borderLeft: '16px solid #eeeeee',
                  pointerEvents: 'none'
                }}
              />
            </>

          )}
        </ButtonGroup>
        <div style={{ display: 'flex', marginLeft: 'auto', padding: '0 10px' }}>
          <h3
            className='h3-finished-tasks-indicator'
            style={{
              textTransform: 'uppercase',
              lineHeight: '33px',
              margin: 0,
              fontWeight: 800,
              fontSize: '13px',
              letterSpacing: '2px',
              justifySelf: 'flex-end'
            }}
          >Finished Tasks &middot; <span style={{ color: Colors.GREEN5 }}>{activePipeline.finishedTasks}</span>
            <em style={{ color: '#dddddd', padding: '0 4px', textTransform: 'lowercase' }}>/</em>{activePipeline.totalTasks}
          </h3>
          {/* <span style={{fontSize: '16px', fontWeight: 700, marginLeft: 'auto', lineHeight: '33px'}} /> */}
          {/* {Number((activePipeline.finishedTasks / activePipeline.totalTasks) * 100).toFixed(1)}% */}
        </div>
      </Card>
      {/* </CSSTransition> */}
    </>
  )
}

export default StagePanel
