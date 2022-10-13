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

import React, { useState, useEffect } from 'react'
import { Dialog, Button, Toaster, Position, Intent } from '@blueprintjs/core'
import { CopyToClipboard } from 'react-copy-to-clipboard'

import { ReactComponent as CopyIcon } from '@/images/icons/copy.svg'
import * as S from './styled'

const CopyToaster = Toaster.create({
  position: Position.TOP_RIGHT
})

export const ViewOrEditModal = ({ record, onSubmit, onCancel }) => {
  const [name, setName] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    setName(record?.name)
  }, [record])

  const handleCancel = () => {
    setName('')
    setError('')
    onCancel()
  }

  const handleInputChange = (e) => {
    setName(e.target.value)
    setError('')
  }

  const handleSubmit = () => {
    if (!name) {
      setError('Name is required')
      return
    }

    onSubmit(record.id, { name })
    handleCancel()
  }

  return (
    <Dialog
      isOpen={true}
      title='View or Edit Incoming Webhook'
      style={{ width: 640 }}
      onClose={handleCancel}
    >
      <S.FormWrapper>
        <div className='form'>
          <h2>Incoming Webhook Name *</h2>
          <p>
            Give your Incoming Webhook a unique name to help you identify it in
            the future.
          </p>
          <input
            type='text'
            placeholder='Your Incoming Webhook Name'
            className={error ? 'has-error' : ''}
            value={name || ''}
            onChange={handleInputChange}
          />
          {error && <p className='error'>{error}</p>}
        </div>
        <div className='url'>
          <h2>POST URL</h2>
          <p>
            Copy the following URLs to your issue tracking tool for Incidents
            and CI tool for Deployments by making a POST to DevLake.
          </p>
          <h3>Incident</h3>
          <p>POST to register an incident </p>
          <div className='block'>
            <span>{record.postIssuesEndpoint}</span>
            <CopyToClipboard
              text={record.postIssuesEndpoint}
              onCopy={() =>
                CopyToaster.show({
                  message: 'Copy successfully.',
                  intent: Intent.SUCCESS
                })
              }
            >
              <CopyIcon width={16} height={16} />
            </CopyToClipboard>
          </div>
          <p>POST to close a registered incident</p>
          <div className='block'>
            <span>{record.closeIssuesEndpoint}</span>
            <CopyToClipboard
              text={record.closeIssuesEndpoint}
              onCopy={() =>
                CopyToaster.show({
                  message: 'Copy successfully.',
                  intent: Intent.SUCCESS
                })
              }
            >
              <CopyIcon width={16} height={16} />
            </CopyToClipboard>
          </div>
          <h3>Deployment</h3>
          <p>POST to register a deployment</p>
          <div className='block'>
            <span>{record.postDeploymentsCurl}</span>
            <CopyToClipboard text={record.postDeploymentsCurl}>
              <CopyIcon width={16} height={16} />
            </CopyToClipboard>
          </div>
        </div>
        <div className='btns'>
          <Button onClick={onCancel}>Cancel</Button>
          <Button intent={Intent.PRIMARY} onClick={handleSubmit}>
            Save
          </Button>
        </div>
      </S.FormWrapper>
    </Dialog>
  )
}
