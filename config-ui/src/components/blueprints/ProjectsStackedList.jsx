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
import {
  Button,
  Icon,
  Intent,
  Elevation,
  Card,
  Colors,
} from '@blueprintjs/core'

const ProjectsStackedList = (props) => {
  const {
    projects = [],
    configuredConnection,
    configuredProject,
    addProjectTransformation = () => {},
  } = props

  return (
    <>
      {projects[configuredConnection.id]?.length > 0 && (
        <Card
          className='selected-connections-list'
          elevation={Elevation.ZERO}
          style={{ padding: 0, marginTop: '10px' }}
        >
          {projects[configuredConnection.id]?.map((project, pIdx) => (
            <div
              className='project-entry'
              key={`project-row-key-${pIdx}`}
              style={{
                display: 'flex',
                width: '100%',
                height: '32px',
                lineHeight: '100%',
                justifyContent: 'space-between',
                // margin: '8px 0',
                padding: '8px 12px',
                borderBottom: '1px solid #f0f0f0',
                backgroundColor:
                  configuredProject === project
                    ? 'rgba(116, 151, 247, 0.2)'
                    : '#fff',
              }}
            >
              <div>
                <div className='project-name' style={{ fontWeight: 600 }}>
                  <input
                    type='radio'
                    name='configured-project'
                    checked={project === configuredProject}
                    onChange={() => addProjectTransformation(project)}
                  />{' '}
                  {project}
                </div>
              </div>
              <div
                style={{
                  display: 'flex',
                  alignContent: 'center',
                }}
              >
                <div
                  className='connection-actions'
                  style={{ paddingLeft: '20px' }}
                >
                  <Button
                    intent={Intent.PRIMARY}
                    className='project-action-transformation'
                    icon={
                      <Icon
                        // icon='plus'
                        size={12}
                        color={Colors.BLUE4}
                      />
                    }
                    text='Add Transformation'
                    color={Colors.BLUE3}
                    small
                    minimal={configuredProject !== project ? true : false}
                    style={{
                      minWidth: '18px',
                      minHeight: '18px',
                      fontSize: '11px',
                    }}
                    onClick={() => addProjectTransformation(project)}
                  />
                </div>
              </div>
            </div>
          ))}
        </Card>
      )}
    </>
  )
}

export default ProjectsStackedList
