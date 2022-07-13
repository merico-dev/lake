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
import React, { useEffect, useState } from 'react'
import { useHistory } from 'react-router-dom'

import '@/styles/integration.scss'
import '@/styles/connections.scss'

export default function GitlabSettings (props) {
  const { connection, transformation = {}, provider, onSettingsChange = () => {} } = props
  const history = useHistory()

  useEffect(() => {
    const settings = {
      // no additional settings
    }
    onSettingsChange(settings)
    console.log('>> GITLAB INSTANCE SETTINGS FIELDS CHANGED!', settings)
  }, [
    onSettingsChange
  ])

  const cancel = () => {
    history.push(`/integrations/${provider.id}`)
  }

  useEffect(() => {
    if (connection && connection.ID) {
      console.log('>> GITLAB CONNECTION OBJECT RECEIVED...', connection)
    } else {
      console.log('>>>> WARNING!! NO CONNECTION OBJECT', connection)
    }
  }, [connection])

  return (
    <>
      <div className='headlineContainer'>
        <h5>No Additional Settings</h5>
        <p className='description'>
          This project doesn’t require any configuration.
        </p>
      </div>
    </>
  )
}
