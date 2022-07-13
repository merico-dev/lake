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
  Button,
  ButtonGroup,
  Colors,
  FormGroup,
  Icon,
  InputGroup,
  Intent,
  Label,
  MenuItem,
  Position,
  Tag,
} from '@blueprintjs/core'
import { MultiSelect, Select } from '@blueprintjs/select'

import '@/styles/integration.scss'
import '@/styles/connections.scss'

const MAPPING_TYPES = {
  Requirement: 'Requirement',
  Incident: 'Incident',
  Bug: 'Bug'
}

const createTypeMapObject = (customType, standardType) => {
  return customType && standardType
    ? {
        [customType]: {
          standardType
        }
      }
    : null
}

export default function JiraSettings (props) {
  const {
    connection,
    configuredBoard,
    // eslint-disable-next-line no-unused-vars
    configuredProject,
    transformation = {},
    isSaving,
    onSettingsChange = () => {},
    // eslint-disable-next-line no-unused-vars
    apiVersion = 2,
    issueTypes = [],
    fields = [],
    // eslint-disable-next-line no-unused-vars
    boards = [],
    jiraProxyError,
    isFetchingJIRA = false
  } = props

  const [typeMappingBug, setTypeMappingBug] = useState([])
  const [typeMappingIncident, setTypeMappingIncident] = useState([])
  const [typeMappingRequirement, setTypeMappingRequirement] = useState([])
  const [typeMappingAll, setTypeMappingAll] = useState({})
  // eslint-disable-next-line no-unused-vars
  const [statusMappings, setStatusMappings] = useState()
  const [jiraIssueEpicKeyField, setJiraIssueEpicKeyField] = useState('')
  const [jiraIssueStoryPointField, setJiraIssueStoryPointField] = useState('')
  // eslint-disable-next-line no-unused-vars
  const [remoteLinkCommitSha, setRemoteLinkCommitSha] = useState('')

  const [requirementTags, setRequirementTags] = useState([])
  const [bugTags, setBugTags] = useState([])
  const [incidentTags, setIncidentTags] = useState([])

  const [requirementTagsList, setRequirementTagsList] = useState([])
  const [bugTagsList, setBugTagsList] = useState([])
  const [incidentTagsList, setIncidentTagsList] = useState([])

  const [fieldsList, setFieldsList] = useState(fields)
  // const [issueTypesList, setIssueTypesList] = useState(issueTypes)

  // useEffect(() => {
  //   onSettingsChange({ ...transformation, typeMappings: typeMappingAll }, configuredBoard?.id)
  // }, [
  //   typeMappingAll,
  //   onSettingsChange,
  //   configuredBoard?.id,
  //   transformation
  // ])

  useEffect(() => {
    if (typeMappingBug && typeMappingIncident && typeMappingRequirement) {
      const RequirementMappings = typeMappingRequirement !== ''
        ? typeMappingRequirement.map(r => createTypeMapObject(r.value, MAPPING_TYPES.Requirement))
        : []
      const IncidentMappings = typeMappingIncident !== ''
        ? typeMappingIncident.map(i => createTypeMapObject(i.value, MAPPING_TYPES.Incident))
        : []
      const BugMappings = typeMappingBug !== ''
        ? typeMappingBug.map(b => createTypeMapObject(b.value, MAPPING_TYPES.Bug))
        : []
      const CombinedMappings = [...RequirementMappings, ...IncidentMappings, ...BugMappings].filter(m => m !== null)
      const MappingTypeObjects = CombinedMappings.reduce((pV, cV) => { return { ...cV, ...pV } }, {})
      setTypeMappingAll(MappingTypeObjects)
      console.log('>> INCIDENT TYPE MAPPING OBJECTS....', RequirementMappings, IncidentMappings, BugMappings)
      console.log('>> ALL MAPPINGS COMBINED...', CombinedMappings)
      console.log('>> FINAL MAPPING OBJECTS FOR API REQUEST...', MappingTypeObjects)
    }
  }, [typeMappingBug, typeMappingIncident, typeMappingRequirement])

  useEffect(() => {
    console.log('>> CONN SETTINGS OBJECT ', connection)
    if (connection && connection.ID) {
      // Parse Type Mappings (V2)
      // setStatusMappings([])
    }
  }, [connection])

  // @todo: FIX Error: Maximum update depth exceeded.
  useEffect(() => {
    setTypeMappingRequirement(requirementTags)
    onSettingsChange({ ...transformation, requirementTags: requirementTags }, configuredBoard?.id)
  }, [requirementTags, configuredBoard?.id, onSettingsChange])

  useEffect(() => {
    setTypeMappingBug(bugTags)
    onSettingsChange({ ...transformation, bugTags: bugTags }, configuredBoard?.id)
  }, [bugTags, configuredBoard?.id, onSettingsChange])

  useEffect(() => {
    setTypeMappingIncident(incidentTags)
    onSettingsChange({ ...transformation, incidentTags: incidentTags }, configuredBoard?.id)
  }, [incidentTags, configuredBoard?.id, onSettingsChange])

  useEffect(() => {
    console.log('>>> JIRA SETTINGS :: FIELDS LIST DATA CHANGED!', fields)
    setFieldsList(fields)
  }, [fields])

  useEffect(() => {
    console.log('>>> JIRA SETTINGS :: ISSUE TYPES LIST DATA CHANGED!', issueTypes)
    setRequirementTagsList(issueTypes)
    setBugTagsList(issueTypes)
    setIncidentTagsList(issueTypes)
  }, [issueTypes])

  useEffect(() => {
    setJiraIssueEpicKeyField(fieldsList.find(f => f.value === transformation?.epicKeyField))
  }, [fieldsList, transformation?.epicKeyField])

  useEffect(() => {
    setJiraIssueStoryPointField(fieldsList.find(f => f.value === transformation?.storyPointField))
  }, [fieldsList, transformation?.storyPointField])

  useEffect(() => {
    console.log('>>>> TRANSFORMATION SETTINGS OBJECT....', transformation)
    setRemoteLinkCommitSha(transformation?.remotelinkCommitShaPattern)
  }, [transformation])

  useEffect(() => {
    console.log('>>>> CONFIGURING BOARD....', configuredBoard)
  }, [configuredBoard])

  return (
    <>
      <h5>Issue Tracking</h5>
      <p className=''>Map your issue labels with each category
        to view corresponding metrics in the
        dashboard.
      </p>

      <div className='issue-type-multiselect' style={{ display: 'flex', marginBottom: '10px' }}>
        <div className='issue-type-label' style={{ minWidth: '120px', paddingRight: '10px', paddingTop: '3px' }}>
          <label>Requirement</label>
        </div>
        <div className='issue-type-multiselect-selector' style={{ minWidth: '200px', width: '100%' }}>
          <MultiSelect
            disabled={isSaving}
            resetOnSelect={true}
            placeholder='Select...'
            popoverProps={{ usePortal: false, popoverClassName: 'transformation-select-popover', minimal: true, fill: true, style: { width: '100%' } }}
            className='multiselector-requirement-type'
            inline={true}
            fill={true}
            items={requirementTagsList}
            selectedItems={transformation?.requirementTags}
            activeItem={null}
            itemPredicate={(query, item) => item?.title.toLowerCase().indexOf(query.toLowerCase()) >= 0}
            itemRenderer={(item, { handleClick, modifiers }) => (
              <MenuItem
                active={modifiers.active || requirementTags.includes(item)}
                disabled={requirementTags.includes(item)}
                key={item.value}
                label={<span style={{ marginLeft: '20px' }}>{item.description || item.value}</span>}
                onClick={handleClick}
                text={requirementTags.includes(item)
                  ? (
                    <>
                      <img src={item.iconUrl} width={12} height={12} /> {item.title} <Icon icon='small-tick' color={Colors.GREEN5} />
                    </>
                    )
                  : (
                    <span style={{ fontWeight: 700 }}>
                      <img src={item.iconUrl} width={12} height={12} /> {item.title}
                    </span>
                    )}
                style={{ marginBottom: '2px', fontWeight: requirementTags.includes(item) ? 700 : 'normal' }}
              />
            )}
            tagRenderer={(item) => item.title}
            tagInputProps={{
              tagProps: {
                intent: Intent.NONE,
                color: Colors.RED3,
                minimal: true
              },
            }}
            noResults={<MenuItem disabled={true} text='No results.' />}
            onRemove={(item) => {
              setRequirementTags((rT) => rT.filter(t => t.id !== item.id))
            }}
            onItemSelect={(item) => {
              setRequirementTags((rT) => !rT.includes(item) ? [...rT, item] : [...rT])
            }}
          />
        </div>
        <div className='multiselect-clear-action' style={{ marginLeft: '0' }}>
          <Button
            icon='eraser'
            disabled={requirementTags.length === 0 || isSaving}
            intent={Intent.NONE} minimal={false} onClick={() => setRequirementTags([])}
            style={{ borderTopLeftRadius: 0, borderBottomLeftRadius: 0, marginLeft: '-2px' }}
          />
        </div>
      </div>

      <div className='issue-type-multiselect' style={{ display: 'flex', marginBottom: '10px' }}>
        <div className='issue-type-label' style={{ minWidth: '120px', paddingRight: '10px', paddingTop: '3px' }}>
          <label>Bug</label>
        </div>
        <div className='issue-type-multiselect-selector' style={{ minWidth: '200px', width: '100%' }}>
          <MultiSelect
            disabled={isSaving}
            resetOnSelect={true}
            placeholder='Select...'
            popoverProps={{ usePortal: false, popoverClassName: 'transformation-select-popover', minimal: true }}
            className='multiselector-bug-type'
            inline={true}
            fill={true}
            items={bugTagsList}
            selectedItems={transformation?.bugTags}
            activeItem={null}
            itemPredicate={(query, item) => item?.title.toLowerCase().indexOf(query.toLowerCase()) >= 0}
            itemRenderer={(item, { handleClick, modifiers }) => (
              <MenuItem
                active={modifiers.active || bugTags.includes(item)}
                disabled={bugTags.includes(item)}
                key={item.value}
                label={<span style={{ marginLeft: '20px' }}>{item.description || item.value}</span>}
                onClick={handleClick}
                text={bugTags.includes(item)
                  ? (
                    <>
                      <img src={item.iconUrl} width={12} height={12} /> {item.title} <Icon icon='small-tick' color={Colors.GREEN5} />
                    </>
                    )
                  : (
                    <span style={{ fontWeight: 700 }}>
                      <img src={item.iconUrl} width={12} height={12} /> {item.title}
                    </span>
                    )}
                style={{ marginBottom: '2px', fontWeight: bugTags.includes(item) ? 700 : 'normal' }}
              />
            )}
            tagRenderer={(item) => item.title}
            tagInputProps={{
              tagProps: {
                intent: Intent.NONE,
                color: Colors.RED3,
                minimal: true
              },
            }}
            noResults={<MenuItem disabled={true} text='No results.' />}
            onRemove={(item) => {
              setBugTags((rT) => rT.filter(t => t.id !== item.id))
            }}
            onItemSelect={(item) => {
              setBugTags((rT) => !rT.includes(item) ? [...rT, item] : [...rT])
            }}
          />
        </div>
        <div className='multiselect-clear-action' style={{ marginLeft: '0' }}>
          <Button
            icon='eraser'
            disabled={bugTags.length === 0 || isSaving}
            intent={Intent.NONE} minimal={false} onClick={() => setBugTags([])}
            style={{ borderTopLeftRadius: 0, borderBottomLeftRadius: 0, marginLeft: '-2px' }}
          />
        </div>
      </div>

      <div className='issue-type-multiselect' style={{ display: 'flex', marginBottom: '10px' }}>
        <div className='issue-type-label' style={{ minWidth: '120px', paddingRight: '10px', paddingTop: '3px' }}>
          <label>Incident</label>
        </div>
        <div className='issue-type-multiselect-selector' style={{ minWidth: '200px', width: '100%' }}>
          <MultiSelect
            disabled={isSaving}
            resetOnSelect={true}
            placeholder='Select...'
            popoverProps={{ usePortal: false, popoverClassName: 'transformation-select-popover', minimal: true }}
            className='multiselector-incident-type'
            inline={true}
            fill={true}
            items={incidentTagsList}
            selectedItems={transformation?.incidentTags}
            activeItem={null}
            itemPredicate={(query, item) => item?.title.toLowerCase().indexOf(query.toLowerCase()) >= 0}
            itemRenderer={(item, { handleClick, modifiers }) => (
              <MenuItem
                active={modifiers.active || incidentTags.includes(item)}
                disabled={incidentTags.includes(item)}
                key={item.value}
                label={<span style={{ marginLeft: '20px' }}>{item.description || item.value}</span>}
                onClick={handleClick}
                text={incidentTags.includes(item)
                  ? (
                    <>
                      <img src={item.iconUrl} width={12} height={12} /> {item.title} <Icon icon='small-tick' color={Colors.GREEN5} />
                    </>
                    )
                  : (
                    <span style={{ fontWeight: 700 }}>
                      <img src={item.iconUrl} width={12} height={12} /> {item.title}
                    </span>
                    )}
                style={{ marginBottom: '2px', fontWeight: incidentTags.includes(item) ? 700 : 'normal' }}
              />
            )}
            tagRenderer={(item) => item.title}
            tagInputProps={{
              tagProps: {
                intent: Intent.NONE,
                color: Colors.RED3,
                minimal: true
              },
            }}
            noResults={<MenuItem disabled={true} text='No results.' />}
            onRemove={(item) => {
              setIncidentTags((rT) => rT.filter(t => t.id !== item.id))
            }}
            onItemSelect={(item) => {
              setIncidentTags((rT) => !rT.includes(item) ? [...rT, item] : [...rT])
            }}
          />
        </div>
        <div className='multiselect-clear-action' style={{ marginLeft: '0' }}>
          <Button
            icon='eraser'
            disabled={incidentTags.length === 0 || isSaving}
            intent={Intent.NONE} minimal={false} onClick={() => setIncidentTags([])}
            style={{ borderTopLeftRadius: 0, borderBottomLeftRadius: 0, marginLeft: '-2px' }}
          />
        </div>
      </div>

      <div className='epic-key-select' style={{ display: 'flex', marginBottom: '10px' }}>
        <div className='epick-key-label' style={{ minWidth: '120px', paddingRight: '10px', paddingTop: '3px' }}>
          <label>Epic Key</label>
        </div>
        <div style={{ display: 'flex', minWidth: '260px' }}>
          <ButtonGroup disabled={isSaving} fill={true}>
            <Select
              disabled={isSaving || fieldsList.length === 0}
              className='select-epic-key'
              inline={true}
              fill={true}
              items={fieldsList}
              activeItem={jiraIssueEpicKeyField}
              itemPredicate={(query, item) => item?.title.toLowerCase().indexOf(query.toLowerCase()) >= 0}
              itemRenderer={(item, { handleClick, modifiers }) => (
                <MenuItem
                  disabled={jiraIssueStoryPointField?.value === item.value}
                  active={false}
                  intent={modifiers.active ? Intent.NONE : Intent.NONE}
                  key={item.value}
                  label={item.value}
                  onClick={handleClick}
                  text={
                    <>
                      <span>{item.title}</span>{' '}
                      <Tag minimal intent={Intent.PRIMARY} style={{ fontSize: '9px' }}>{item.type}</Tag>{' '}
                      {modifiers.active && (<Icon icon='small-tick' color={Colors.GREEN5} size={14} />)}
                    </>
                  }
                  style={{
                    fontSize: '11px',
                    fontWeight: modifiers.active ? 800 : 'normal',
                    backgroundColor: modifiers.active ? Colors.LIGHT_GRAY4 : 'none'
                  }}
                />
              )}
              noResults={<MenuItem disabled={true} text='No epic results.' />}
              onItemSelect={(item) => {
                setJiraIssueEpicKeyField(item)
                onSettingsChange({ ...transformation, epicKeyField: item?.value }, configuredBoard?.id)
              }}
              popoverProps={{
                position: Position.TOP
              }}
            >
              <Button
                intent={Intent.NONE}
                disabled={isSaving || fieldsList.length === 0}
                fill={true}
                style={{ justifyContent: 'space-between', display: 'flex', minWidth: '260px', maxWidth: '100%' }}
                text={jiraIssueEpicKeyField ? `${jiraIssueEpicKeyField.title}` : 'Select...'}
                rightIcon='double-caret-vertical'
                outlined
              />
            </Select>
            <Button
              disabled={!jiraIssueEpicKeyField || isSaving}
              icon='eraser'
              intent={jiraIssueEpicKeyField ? Intent.NONE : Intent.NONE} minimal={false} onClick={() => {
                setJiraIssueEpicKeyField('')
                onSettingsChange({ ...transformation, epicKeyField: '' }, configuredBoard?.id)
              }}
            />
          </ButtonGroup>
          <div style={{ marginLeft: '10px' }}>
            {jiraProxyError && (
              <p style={{ color: Colors.GRAY4 }}>
                <Icon icon='warning-sign' color={Colors.RED5} size={12} style={{ marginBottom: '2px' }} />{' '}
                {jiraProxyError.toString() || 'JIRA API not accessible.'}
              </p>
            )}
          </div>
        </div>
      </div>

      <div className='story-point-select' style={{ display: 'flex', marginBottom: '10px' }}>
        <div className='story-point-label' style={{ minWidth: '120px', paddingRight: '10px', paddingTop: '3px' }}>
          <label>Story Point Field</label>
        </div>
        <div style={{ display: 'flex', minWidth: '260px' }}>
          <ButtonGroup disabled={isSaving}>
            <Select
              disabled={isSaving || fieldsList.length === 0}
              className='select-story-key'
              inline={true}
              fill={true}
              items={fieldsList}
              activeItem={jiraIssueStoryPointField}
              itemPredicate={(query, item) => item?.title.toLowerCase().indexOf(query.toLowerCase()) >= 0}
              itemRenderer={(item, { handleClick, modifiers }) => (
                <MenuItem
                  disabled={jiraIssueEpicKeyField?.value === item.value}
                  active={false}
                  intent={modifiers.active ? Intent.NONE : Intent.NONE}
                  key={item.value}
                  label={item.value}
                  onClick={handleClick}
                  text={
                    <>
                      {item.title}{' '}
                      <Tag minimal intent={Intent.PRIMARY} style={{ fontSize: '9px' }}>{item.type}</Tag>{' '}
                      {modifiers.active && (<Icon icon='small-tick' color={Colors.GREEN5} size={14} />)}
                    </>
                  }
                  style={{
                    fontSize: '11px',
                    fontWeight: modifiers.active ? 800 : 'normal',
                    backgroundColor: modifiers.active ? Colors.LIGHT_GRAY4 : 'none'
                  }}
                />
              )}
              noResults={<MenuItem disabled={true} text='No epic results.' />}
              onItemSelect={(item) => {
                setJiraIssueStoryPointField(item)
                onSettingsChange({ ...transformation, storyPointField: item?.value }, configuredBoard?.id)
              }}
              popoverProps={{
                position: Position.TOP
              }}
            >
              <Button
                // loading={isFetchingJIRA}
                disabled={isSaving || fieldsList.length === 0}
                fill={true}
                style={{ justifyContent: 'space-between', display: 'flex', minWidth: '260px', maxWidth: '300px' }}
                text={jiraIssueStoryPointField ? `${jiraIssueStoryPointField.title}` : 'Select...'}
                rightIcon='double-caret-vertical'
                outlined
              />
            </Select>
            <Button
              loading={isFetchingJIRA}
              disabled={!jiraIssueStoryPointField || isSaving}
              icon='eraser'
              intent={jiraIssueStoryPointField
                ? Intent.NONE
                : Intent.NONE} minimal={false} onClick={() => {
                  setJiraIssueStoryPointField('')
                  onSettingsChange({ ...transformation, storyPointField: '' }, configuredBoard?.id)
                }}
            />
          </ButtonGroup>
          <div style={{ marginLeft: '10px' }}>
            {jiraProxyError && (
              <p style={{ color: Colors.GRAY4 }}>
                <Icon icon='warning-sign' color={Colors.RED5} size={12} style={{ marginBottom: '2px' }} />{' '}
                {jiraProxyError.toString() || 'JIRA API not accessible.'}
              </p>
            )}
          </div>
        </div>
      </div>
      <div className='headlineContainer'>
        <h5>Additional Settings</h5>
        {/* <h3 className='headline'>Remotelink Commit SHA <Tag className='bp3-form-helper-text'>RegExp</Tag></h3> */}
      </div>
      <div className='formContainer' style={{ maxWidth: '550px' }}>
        <FormGroup
          disabled={isSaving}
          // label={<></>}
          labelFor='jira-remotelink-sha'
          className='formGroup'
          contentClassName='formGroupContent'
        >
          <label>Remotelink Commit SHA</label>
          <p>Issue Weblink <strong>Commit SHA Pattern</strong> (Add weblink to jira for gitlab.)</p>
          <InputGroup
            id='jira-remotelink-sha'
            fill={true}
            placeholder='/commit/([0-9a-f]{40})$'
            value={transformation?.remotelinkCommitShaPattern}
            onChange={(e) => onSettingsChange({ ...transformation, remotelinkCommitShaPattern: e.target.value }, configuredBoard?.id)}
            disabled={isSaving}
            className='input'
          />
        </FormGroup>

      </div>
    </>
  )
}
