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
import {
  useParams,
  useHistory
} from 'react-router-dom'

import '@/styles/integration.scss'
import '@/styles/connections.scss'

export default function JenkinsSettings (props) {
  const { provider, onSettingsChange = () => {} } = props
  const history = useHistory()
  const { providerId, connectionId } = useParams()

  const [errors, setErrors] = useState([])

  const cancel = () => {
    history.push(`/integrations/${provider.id}`)
  }

  // useEffect(() => {
  //   setErrors(['This integration doesn’t require any configuration.'])
  // }, [])

  useEffect(() => {
    onSettingsChange({
      errors,
      providerId,
      connectionId
    })
  }, [errors, onSettingsChange, connectionId, providerId])

  return (
    <>
      <div className='headlineContainer'>
        <h3 className='headline'>No Additional Settings</h3>
        <p className='description'>
          This integration doesn’t require any configuration.
          You can continue to&nbsp;
          <a href='#' style={{ textDecoration: 'underline' }} onClick={cancel}>add other data connections</a>&nbsp;
          or trigger collection at the <a href='#' style={{ textDecoration: 'underline' }} onClick={cancel}>previous page</a>.
        </p>
      </div>
    </>
  )
}
