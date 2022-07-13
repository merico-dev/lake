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
import React, { useState } from 'react'
import {
  Button,
  ButtonGroup,
  FormGroup,
  Popover,
  InputGroup,
  Intent,
  Tag,
  Tooltip,
  Colors,
} from '@blueprintjs/core'
import RefDiffTasksMenu from '@/components/menus/RefDiffTasksMenu'

const RefDiffSettings = (props) => {
  const {
    providerId,
    repoId,
    tasks = [],
    pairs = [],
    isRunning,
    isEnabled,
    setRepoId = () => {},
    setTasks = () => {},
    setPairs = () => {},
  } = props

  const [refDiffOldTag, setOldTag] = useState('')
  const [refDiffNewTag, setNewTag] = useState('')

  const handleRefDiffTaskSelect = (e, task) => {
    setTasks(t => t.includes(task.task)
      ? t.filter(t2 => t2 !== task.task)
      : [...new Set([...t, task.task])])
  }

  // eslint-disable-next-line no-unused-vars
  const createRefDiffPair = (oldRef, newRef) => {
    return {
      oldRef,
      newRef
    }
  }

  const addRefDiffPairObject = (oldRef, newRef) => {
    setPairs(pairs => (!pairs.some(p => p.oldRef === oldRef && p.newRef === newRef))
      ? [...pairs, { oldRef, newRef }]
      : [...pairs])
    setNewTag('')
    setOldTag('')
  }

  const removeRefDiffPairObject = (oldRef, newRef) => {
    setPairs(pairs => pairs.filter(p => !(p.oldRef === oldRef && p.newRef === newRef)))
  }

  return (
    <>
      <FormGroup
        disabled={isRunning || !isEnabled(providerId)}
        label={<strong>Repository ID<span className='requiredStar'>*</span></strong>}
        labelInfo={<span style={{ display: 'block' }}>Enter Repo Column ID</span>}
        inline={false}
        labelFor='refdiff-repo-id'
        className=''
        contentClassName=''
        style={{ minWidth: '280px', marginBottom: 'auto', whiteSpace: 'nowrap' }}
        fill
      >
        <InputGroup
          id='refdiff-repo-id'
          disabled={isRunning || !isEnabled(providerId)}
          placeholder='eg. github:GithubRepo:384111310'
          value={repoId}
          onChange={(e) => setRepoId(e.target.value)}
          className='input-refdiff-repo-id'
          autoComplete='off'
          fill={false}
          style={{ marginBottom: '10px' }}
        />
      </FormGroup>
      <FormGroup
        disabled={isRunning || !isEnabled(providerId)}
        label={(
          <strong>Tasks<span className='requiredStar'>*</span>
            <span
              className='badge-count'
              style={{
                opacity: isEnabled(providerId) ? 0.5 : 0.1
              }}
            >{tasks.length}
            </span>
          </strong>
        )}
        labelInfo={<span style={{ display: 'block' }}>Select Tasks to Execute</span>}
        inline={false}
        labelFor='refdiff-tasks'
        className=''
        contentClassName=''
        style={{ marginLeft: '12px', marginRight: '12px', marginBottom: 'auto', whiteSpace: 'nowrap' }}
        fill
      >

        <Popover
          className='provider-tasks-popover'
          disabled={isRunning || !isEnabled(providerId)}
          content={(
            <RefDiffTasksMenu onSelect={handleRefDiffTaskSelect} selected={tasks} />
          )}
          placement='top-center'
        >
          <ButtonGroup disabled={isRunning || !isEnabled(providerId)}>
            <Button
              disabled={isRunning || !isEnabled(providerId)}
              icon='menu'
              text={tasks.length > 0
                ? <>Choose Tasks <Tag intent={Intent.PRIMARY} round>{tasks.length}</Tag></>
                : <>Choose Tasks <Tag intent={Intent.PRIMARY} round>None</Tag></>}
            />
            <Button
              icon='eraser'
              intent={Intent.WARNING}
              disabled={isRunning || !isEnabled(providerId) || tasks.length === 0}
              onClick={() => setTasks([])}
            />
          </ButtonGroup>
        </Popover>
      </FormGroup>
      <div style={{ display: 'flex', flex: 1, width: '100%' }}>
        <FormGroup
          disabled={isRunning || !isEnabled(providerId)}
          label={(
            <strong>Tags<span className='requiredStar'>*</span>
              <span
                className='badge-count'
                style={{
                  opacity: isEnabled(providerId) ? 0.5 : 0.1
                }}
              >{pairs.length}
              </span>
            </strong>
          )}
          labelInfo={<span style={{ display: 'block' }}>Specify tag Ref Pairs</span>}
          inline={false}
          labelFor='refdiff-pair-newref'
          className=''
          contentClassName=''
          style={{ minWidth: '582px', marginBottom: 'auto' }}
          fill={false}
        >
          <div style={{ display: 'flex' }}>
            <div>
              <InputGroup
                id='refdiff-pair-newref'
                leftElement={(
                  <Tag
                    intent={Intent.WARNING} style={{
                      opacity: isEnabled(providerId) ? 1 : 0.3
                    }}
                  >New Ref
                  </Tag>
                )}
                inline={true}
                disabled={isRunning || !isEnabled(providerId)}
                placeholder='eg. refs/tags/v0.6.0'
                value={refDiffNewTag}
                onChange={(e) => setNewTag(e.target.value)}
                autoComplete='off'
                fill={false}
              />
            </div>
            <div style={{ marginLeft: '10px', marginRight: 0 }}>
              <InputGroup
                id='refdiff-pair-oldref'
                leftElement={(
                  <Tag
                    style={{
                      opacity: isEnabled(providerId) ? 1 : 0.3
                    }}
                  >Old Ref
                  </Tag>
                )}
                inline={true}
                disabled={isRunning || !isEnabled(providerId)}
                placeholder='eg. refs/tags/v0.5.0'
                value={refDiffOldTag}
                onChange={(e) => setOldTag(e.target.value)}
                autoComplete='off'
                fill={false}
                rightElement={(
                  <Tooltip content='Add Tag Pair'>
                    <Button
                      className='btn-add-tagpair'
                      intent={Intent.PRIMARY}
                      disabled={!refDiffOldTag || !refDiffNewTag || refDiffOldTag === refDiffNewTag}
                      icon='plus'
                      style={{}}
                      onClick={() => addRefDiffPairObject(refDiffOldTag, refDiffNewTag)}
                    />
                  </Tooltip>
                )}
              />
            </div>
            <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
              {/* <Button icon='remove' minimal small style={{ marginTop: 'auto', alignSelf: 'center' }} /> */}
              {/* <Button
              intent={Intent.PRIMARY}
              disabled={!refDiffOldTag || !refDiffNewTag || refDiffOldTag === refDiffNewTag}
              icon='add'
              small
              style={{ marginTop: 'auto', alignSelf: 'center' }}
              onClick={() => addRefDiffPairObject(refDiffOldTag, refDiffNewTag)}
            /> */}
            </div>
          </div>
          <div
            style={{
              borderRadius: '4px',
              padding: '10px',
              margin: '5px 0'
            }}
          >
            {pairs.map((pair, pairIdx) => (
              <div key={`refdiff-added-pairs-itemkey-$${pairIdx}`} style={{ display: 'flex' }}>
                <div style={{ flex: 1 }}>
                  <Tag intent={Intent.WARNING} round='false' small>new</Tag> {pair.newRef}
                </div>
                <div style={{ flex: 1, marginLeft: '10px', marginRight: '10px' }}>
                  <Tag round='false' small>old</Tag> {pair.oldRef}
                </div>
                <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                  <Button
                    className='btn-remove-tagpair'
                    icon='remove'
                    minimal
                    small='true'
                    style={{ marginTop: 'auto', alignSelf: 'center' }}
                    onClick={() => removeRefDiffPairObject(pair.oldRef, pair.newRef)}
                  />
                </div>
              </div>
            ))}
            {pairs.length === 0 && <><span style={{ color: Colors.GRAY3 }}>( No Tag Pairs Added)</span></>}
          </div>
        </FormGroup>
      </div>
    </>
  )
}

export default RefDiffSettings
