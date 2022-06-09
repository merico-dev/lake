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
import { useParams } from 'react-router-dom'
import { FormGroup, InputGroup, Label, Tag } from '@blueprintjs/core'

import '@/styles/integration.scss'
import '@/styles/connections.scss'

export default function GithubSettings (props) {
  const { connection, isSaving, isSavingConnection, onSettingsChange = () => {} } = props
  const { providerId, connectionId } = useParams()
  const [prType, setPrType] = useState('')
  const [prComponent, setPrComponent] = useState('')
  const [issueSeverity, setIssueSeverity] = useState('')
  const [issueComponent, setIssueComponent] = useState('')
  const [issuePriority, setIssuePriority] = useState('')
  const [issueTypeBug, setIssueTypeBug] = useState('')
  const [issueTypeRequirement, setIssueTypeRequirement] = useState('')
  const [issueTypeIncident, setIssueTypeIncident] = useState('')

  const [errors, setErrors] = useState([])

  useEffect(() => {
    setErrors(['This integration doesn’t require any configuration.'])
  }, [])

  useEffect(() => {
    onSettingsChange({
      errors,
      providerId,
      connectionId
    })
  }, [errors, onSettingsChange, connectionId, providerId])

  useEffect(() => {
    setPrType(connection.prType)
    setPrComponent(connection.prComponent)
    setIssueSeverity(connection.issueSeverity)
    setIssuePriority(connection.issuePriority)
    setIssueComponent(connection.issueComponent)
    setIssueTypeBug(connection.issueTypeBug)
    setIssueTypeRequirement(connection.issueTypeRequirement)
    setIssueTypeIncident(connection.issueTypeIncident)
  }, [connection])

  useEffect(() => {
    const settings = {
      prType: prType,
      prComponent: prComponent,
      issueSeverity: issueSeverity,
      issueComponent: issueComponent,
      issuePriority: issuePriority,
      issueTypeRequirement: issueTypeRequirement,
      issueTypeBug: issueTypeBug,
      issueTypeIncident: issueTypeIncident,
    }
    console.log('>> GITHUB INSTANCE SETTINGS FIELDS CHANGED!', settings)
    onSettingsChange(settings)
  }, [
    prType,
    prComponent,
    issueSeverity,
    issueComponent,
    issuePriority,
    issueTypeRequirement,
    issueTypeBug,
    issueTypeIncident,
    onSettingsChange
  ])

  return (
    <>
      <h5>Issue Tracking{' '} <Tag className='bp3-form-helper-text'>RegExp</Tag></h5>
      <p className=''>Map your issue labels with each category
        to view corresponding metrics in the
        dashboard.</p>
      <div style={{ maxWidth: '60%' }}>
        <div className='formContainer'>
          <FormGroup
            disabled={isSaving || isSavingConnection}
            inline={true}
            label='Severity'
            labelFor='github-issue-severity'
            className='formGroup'
            contentClassName='formGroupContent'
          >
            <InputGroup
              id='github-issue-severity'
              placeholder='severity/(.*)$'
              defaultValue={issueSeverity}
              onChange={(e) => setIssueSeverity(e.target.value)}
              onKeyUp={(e) => e.target.value.length === 0 ? setIssueSeverity('') : null}
              disabled={isSaving || isSavingConnection}
              className='input'
              maxLength={255}
            />
          </FormGroup>
        </div>
        <div className='formContainer'>
          <FormGroup
            disabled={isSaving || isSavingConnection}
            inline={true}
            label='Component'
            labelFor='github-issue-component'
            className='formGroup'
            contentClassName='formGroupContent'
          >
            <InputGroup
              id='github-issue-component'
              placeholder='component/(.*)$'
              defaultValue={issueComponent}
              onChange={(e) => setIssueComponent(e.target.value)}
              onKeyUp={(e) => e.target.value.length === 0 ? setIssueComponent('') : null}
              disabled={isSaving || isSavingConnection}
              className='input'
              maxLength={255}
            />
          </FormGroup>
        </div>
        <div className='formContainer'>
          <FormGroup
            disabled={isSaving || isSavingConnection}
            inline={true}
            label='Priority'
            labelFor='github-issue-priority'
            className='formGroup'
            contentClassName='formGroupContent'
          >
            <InputGroup
              id='github-issue-priority'
              placeholder='(highest|high|medium|low)$'
              defaultValue={issuePriority}
              onChange={(e) => setIssuePriority(e.target.value)}
              onKeyUp={(e) => e.target.value.length === 0 ? setIssuePriority('') : null}
              disabled={isSaving || isSavingConnection}
              className='input'
              maxLength={255}
            />
          </FormGroup>
        </div>
        <div className='formContainer'>
          <FormGroup
            disabled={isSaving || isSavingConnection}
            inline={true}
            label='Type/Requirement'
            labelFor='github-issue-requirement'
            className='formGroup'
            contentClassName='formGroupContent'
          >
            <InputGroup
              id='github-issue-requirement'
              placeholder='(feat|feature|proposal|requirement)$'
              defaultValue={issueTypeRequirement}
              onChange={(e) => setIssueTypeRequirement(e.target.value)}
              onKeyUp={(e) => e.target.value.length === 0 ? setIssueTypeRequirement('') : null}
              disabled={isSaving || isSavingConnection}
              className='input'
              maxLength={255}
            />
          </FormGroup>
        </div>
        <div className='formContainer'>
          <FormGroup
            disabled={isSaving || isSavingConnection}
            inline={true}
            label='Type/Bug'
            labelFor='github-issue-bug'
            className='formGroup'
            contentClassName='formGroupContent'
          >
            <InputGroup
              id='github-issue-bug'
              placeholder='(bug|broken)$'
              defaultValue={issueTypeBug}
              onChange={(e) => setIssueTypeBug(e.target.value)}
              onKeyUp={(e) => e.target.value.length === 0 ? setIssueTypeBug('') : null}
              disabled={isSaving || isSavingConnection}
              className='input'
              maxLength={255}
            />
          </FormGroup>
        </div>
        <div className='formContainer'>
          <FormGroup
            disabled={isSaving || isSavingConnection}
            inline={true}
            label='Type/Incident'
            labelFor='github-issue-bug'
            className='formGroup'
            contentClassName='formGroupContent'
          >
            <InputGroup
              id='github-issue-incident'
              placeholder='(incident|p0|p1|p2)$'
              defaultValue={issueTypeIncident}
              onChange={(e) => setIssueTypeIncident(e.target.value)}
              onKeyUp={(e) => e.target.value.length === 0 ? setIssueTypeIncident('') : null}
              disabled={isSaving || isSavingConnection}
              className='input'
              maxLength={255}
            />
          </FormGroup>
        </div>
      </div>

      <h5>Code Review{' '} <Tag className='bp3-form-helper-text'>RegExp</Tag></h5>
      <p className=''>Map your pull requests labels with each category to view corresponding metrics in the dashboard.</p>

      <div style={{ maxWidth: '60%' }}>
        <div className='formContainer'>
          <FormGroup
            disabled={isSaving || isSavingConnection}
            inline={true}
            label='Type'
            labelFor='github-pr-type'
            className='formGroup'
            contentClassName='formGroupContent'
          >
            <InputGroup
              id='github-pr-type'
              placeholder='type/(.*)$'
              defaultValue={prType}
              onChange={(e) => setPrType(e.target.value)}
              onKeyUp={(e) => e.target.value.length === 0 ? setPrType('') : null}
              disabled={isSaving || isSavingConnection}
              className='input'
              maxLength={255}
            />
          </FormGroup>
        </div>
        <div className='formContainer'>
          <FormGroup
            disabled={isSaving || isSavingConnection}
            inline={true}
            label='Component'
            labelFor='github-pr-component'
            className='formGroup'
            contentClassName='formGroupContent'
          >
            <InputGroup
              id='github-pr-type'
              placeholder='component/(.*)$'
              defaultValue={prComponent}
              onChange={(e) => setPrComponent(e.target.value)}
              onKeyUp={(e) => e.target.value.length === 0 ? setPrComponent('') : null}
              disabled={isSaving || isSavingConnection}
              className='input'
              maxLength={255}
            />
          </FormGroup>
        </div>
      </div>
    </>
  )
}
