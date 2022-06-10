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
import React, { Fragment } from 'react'
import { Button, Intent, Icon } from '@blueprintjs/core'

const NoData = (props) => {
  const {
    title = 'No Data',
    message = 'Please check configuration and retry.',
    icon = 'offline',
    iconSize = 32,
    actionText = 'Go Back',
    onClick = () => {},
  } = props

  return (
    <>
      <div className='bp3-non-ideal-state no-data'>
        <div className='bp3-non-ideal-state-visual'>
          <Icon icon={icon} size={iconSize} />
        </div>
        <div className='bp3-non-ideal-state-text'>
          <h4 className='bp3-heading' style={{ margin: 0 }}>
            {title}
          </h4>
          <div>{message}</div>
        </div>
        {onClick && actionText && (
          <Button intent={Intent.NONE} onClick={onClick}>
            {actionText}
          </Button>
        )}
      </div>
    </>
  )
}

export default NoData
