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
import { Providers, } from '@/data/Providers'
import { Button, ButtonGroup, Colors, FormGroup, InputGroup, Intent, MenuItem, TagInput, } from '@blueprintjs/core'
import { Select } from '@blueprintjs/select'
import RefDiffSettings from '@/components/pipelines/pipeline-settings/refdiff'

const ProviderSettings = (props) => {
  const {
    providerId,
    projectId = [],
    // connectionId,
    selectedConnection,
    selectedGithubRepo,
    connections = [],
    repositories = [],
    boardId = [],
    owner,
    repositoryName,
    gitExtractorUrl,
    // gitExtractorRepoId,
    refDiffRepoId,
    refDiffPairs = [],
    refDiffTasks = [],
    setProjectId = () => {},
    // setConnectionId = () => {},
    setSelectedConnection = () => {},
    setBoardId = () => {},
    setOwner = () => {},
    setRepositoryName = () => {},
    setGitExtractorUrl = () => {},
    // setGitExtractorRepoId = () => {},
    setSelectedGithubRepo = () => {},
    setRefDiffRepoId = () => {},
    setRefDiffPairs = () => {},
    setRefDiffTasks = () => {},
    isEnabled = () => {},
    isRunning = false,
    // onReset = () => {}
  } = props

  let providerSettings = null

  switch (providerId) {
    case Providers.JENKINS:
      providerSettings = <p><strong style={{ fontWeight: 900 }}>AUTO-CONFIGURED</strong><br />No Additional Settings</p>
      break
    case Providers.JIRA:
      providerSettings = (
        <>
          <FormGroup
            disabled={isRunning || !isEnabled(providerId)}
            label={<strong>Connection ID<span className='requiredStar'>*</span></strong>}
            labelInfo={<span style={{ display: 'block' }}>Choose Connection Instance ID</span>}
            inline={false}
            labelFor='source-id'
            className=''
            contentClassName=''
            fill
            style={{ marginRight: '12px' }}
          >
            <ButtonGroup>
              <Select
                disabled={isRunning || !isEnabled(providerId)}
                className='selector-connection-id'
                popoverProps={{ popoverClassName: 'source-id-popover' }}
                multiple
                inline={true}
                fill={true}
                items={connections}
                activeItem={selectedConnection}
                itemPredicate={(query, item) => item?.title?.toLowerCase().indexOf(query.toLowerCase()) >= 0}
                itemRenderer={(item, { handleClick, modifiers }) => (
                  <MenuItem
                    active={modifiers.active}
                    key={item.value}
                    label={item.value}
                    onClick={handleClick}
                    text={item.title}
                  />
                )}
                noResults={<MenuItem disabled={true} text='No Connections.' />}
                onItemSelect={(item) => {
                  setSelectedConnection(item)
                }}
              >
                <Button
                  className='btn-connection-id-selector'
                  disabled={isRunning || !isEnabled(providerId)}
                  style={{ justifyContent: 'space-between', minWidth: '206px', maxWidth: '290px', whiteSpace: 'nowrap' }}
                  text={selectedConnection ? `${selectedConnection.title} [${selectedConnection.value}]` : 'Select Instance'}
                  rightIcon='double-caret-vertical'
                  fill
                />
              </Select>
              <Button
                icon='eraser'
                intent={Intent.WARNING}
                disabled={isRunning || !isEnabled(providerId)}
                onClick={() => setSelectedConnection(null)}
              />
            </ButtonGroup>
          </FormGroup>
          <FormGroup
            disabled={isRunning || !isEnabled(providerId)}
            label={
              <strong style={{ marginTop: '10px', display: 'inline-block' }}>Board ID<span className='requiredStar'>*</span>
                <span
                  className='badge-count'
                  style={{
                    opacity: isEnabled(providerId) ? 0.5 : 0.1
                  }}
                >{boardId.length}
                </span>
              </strong>
          }
            labelInfo={<span style={{ display: 'block' }}>Enter JIRA Board ID.</span>}
            inline={false}
            labelFor='board-id'
            className=''
            contentClassName=''
            fill
          >
            <div style={{ width: '100%' }}>
              <TagInput
                id='board-id'
                disabled={isRunning || !isEnabled(providerId)}
                placeholder='eg. 8'
                values={boardId || []}
                fill={true}
                onChange={(values) => setBoardId([...new Set(values)])}
                addOnPaste={true}
                addOnBlur={true}
                rightElement={
                  <Button
                    disabled={isRunning || !isEnabled(providerId)}
                    icon='eraser'
                    minimal
                    onClick={() => setBoardId([])}
                  />
              }
                onKeyDown={e => e.key === 'Enter' && e.preventDefault()}
                className='input-board-id tagInput'
              />
            </div>
          </FormGroup>
        </>
      )
      break
    case Providers.GITHUB:
      providerSettings = (
        <>
          <FormGroup
            disabled={isRunning || !isEnabled(providerId)}
            label={<strong>Owner<span className='requiredStar'>*</span></strong>}
            labelInfo={<span style={{ display: 'block' }}>Enter Project Owner</span>}
            inline={false}
            labelFor='owner'
            className=''
            contentClassName=''
            fill
          >
            <InputGroup
              id='owner'
              disabled={isRunning || !isEnabled(providerId)}
              placeholder='eg. merico-dev'
              value={owner}
              onChange={(e) => setOwner(e.target.value)}
              className='input-owner'
              autoComplete='off'
            />
          </FormGroup>
          <FormGroup
            disabled={isRunning || !isEnabled(providerId)}
            label={<strong>Repository Name<span className='requiredStar'>*</span></strong>}
            labelInfo={<span style={{ display: 'block' }}>Enter Git repository</span>}
            inline={false}
            labelFor='repository-name'
            className=''
            contentClassName=''
            style={{ marginLeft: '12px' }}
            fill
          >
            <InputGroup
              id='repository-name'
              disabled={isRunning || !isEnabled(providerId)}
              placeholder='eg. lake'
              value={repositoryName}
              onChange={(e) => setRepositoryName(e.target.value)}
              className='input-repository-name'
              autoComplete='off'
              fill={false}
            />
          </FormGroup>
        </>
      )
      break
    case Providers.GITLAB:
      providerSettings = (
        <>
          <FormGroup
            disabled={isRunning || !isEnabled(providerId)}
            label={<strong>Connection ID<span className='requiredStar'>*</span></strong>}
            labelInfo={<span style={{ display: 'block' }}>Choose Connection Instance ID</span>}
            inline={false}
            labelFor='source-id'
            className=''
            contentClassName=''
            fill
            style={{ marginRight: '12px' }}
          >
            <ButtonGroup>
              <Select
                disabled={isRunning || !isEnabled(providerId)}
                className='selector-connection-id'
                popoverProps={{ popoverClassName: 'source-id-popover' }}
                multiple
                inline={true}
                fill={true}
                items={connections}
                activeItem={selectedConnection}
                itemPredicate={(query, item) => item?.title?.toLowerCase().indexOf(query.toLowerCase()) >= 0}
                itemRenderer={(item, { handleClick, modifiers }) => (
                  <MenuItem
                    active={modifiers.active}
                    key={item.value}
                    label={item.value}
                    onClick={handleClick}
                    text={item.title}
                  />
                )}
                noResults={<MenuItem disabled={true} text='No Connections.' />}
                onItemSelect={(item) => {
                  setSelectedConnection(item)
                }}
              >
                <Button
                  className='btn-connection-id-selector'
                  disabled={isRunning || !isEnabled(providerId)}
                  style={{ justifyContent: 'space-between', minWidth: '206px', maxWidth: '290px', whiteSpace: 'nowrap' }}
                  text={selectedConnection ? `${selectedConnection.title} [${selectedConnection.value}]` : 'Select Instance'}
                  rightIcon='double-caret-vertical'
                  fill
                />
              </Select>
              <Button
                icon='eraser'
                intent={Intent.WARNING}
                disabled={isRunning || !isEnabled(providerId)}
                onClick={() => setSelectedConnection(null)}
              />
            </ButtonGroup>
          </FormGroup>
          <FormGroup
            disabled={isRunning || !isEnabled(providerId)}
            label={
              <strong>Project ID<span className='requiredStar'>*</span>
                <span
                  className='badge-count'
                  style={{
                    opacity: isEnabled(providerId) ? 0.5 : 0.1
                  }}
                >{projectId.length}
                </span>
              </strong>
          }
            labelInfo={<span style={{ display: 'block' }}>Enter GitLab Project ID.</span>}
            inline={false}
            labelFor='project-id'
            className=''
            contentClassName=''
          >
            <div style={{ width: '100%' }}>
              <TagInput
                id='project-id'
                disabled={isRunning || !isEnabled(providerId)}
                placeholder='eg. 937810831'
                values={projectId || []}
                fill={true}
                onChange={(values) => setProjectId([...new Set(values)])}
                addOnPaste={true}
                addOnBlur={true}
                rightElement={
                  <Button
                    disabled={isRunning || !isEnabled(providerId)}
                    icon='eraser'
                    minimal
                    onClick={() => setProjectId([])}
                  />
              }
                onKeyDown={e => e.key === 'Enter' && e.preventDefault()}
                className='input-project-id tagInput'
              />
            </div>
          </FormGroup>
        </>
      )
      break
    case Providers.GITEXTRACTOR:
      providerSettings = (
        <>
          <FormGroup
            disabled={isRunning || !isEnabled(providerId)}
            label={<strong>Git URL<span className='requiredStar'>*</span></strong>}
            labelInfo={<span style={{ display: 'block' }}>Enter Repository URL</span>}
            inline={false}
            labelFor='git-url'
            className=''
            contentClassName=''
            fill
            style={{ minWidth: '372px', marginRight: '12px' }}
          >
            <InputGroup
              id='gitextractor-url'
              disabled={isRunning || !isEnabled(providerId)}
              placeholder='eg. https://github.com/apache/incubator-devlake.git'
              value={gitExtractorUrl}
              onChange={(e) => setGitExtractorUrl(e.target.value)}
              className='input-gitextractor-url'
              autoComplete='off'
            />
          </FormGroup>
          <FormGroup
            disabled={isRunning || !isEnabled(providerId)}
            label={<strong style={{ marginTop: '10px', display: 'inline-block' }}>Repository ID<span className='requiredStar'>*</span></strong>}
            labelInfo={<span style={{ display: 'block' }}>Choose Repo Column ID</span>}
            inline={false}
            labelFor='gitextractor-repo-id'
            className=''
            contentClassName=''
            style={{ minWidth: '280px', whiteSpace: 'nowrap' }}
            fill
          >
            {/* Manual Text Input @DISABLED */}
            {/* <InputGroup
              id='gitextractor-repo-id'
              disabled={isRunning || !isEnabled(providerId)}
              placeholder='eg. github:GithubRepo:384111310'
              value={gitExtractorRepoId}
              onChange={(e) => setGitExtractorRepoId(e.target.value)}
              className='input-gitextractor-repo-id'
              autoComplete='off'
              fill={false}
            /> */}
            <ButtonGroup>
              <Select
                disabled={isRunning || !isEnabled(providerId) || repositories.length === 0}
                className='selector-gitextractor-repo-id'
                popoverProps={{ popoverClassName: 'gitextractor-repo-id-popover' }}
                multiple
                inline={true}
                fill={true}
                items={repositories}
                activeItem={selectedGithubRepo}
                itemPredicate={(query, item) => item?.title?.toLowerCase().indexOf(query.toLowerCase()) >= 0}
                itemRenderer={(item, { handleClick, modifiers }) => (
                  <MenuItem
                    active={modifiers.active}
                    key={item.value}
                    label={item.value}
                    onClick={handleClick}
                    text={item.title}
                  />
                )}
                noResults={<MenuItem disabled={true} text='No Repositories.' />}
                onItemSelect={(item) => {
                  setSelectedGithubRepo(item)
                }}
              >
                <Button
                  className='btn-gitextractor-repo-id-selector'
                  disabled={isRunning || !isEnabled(providerId)}
                  style={{ justifyContent: 'space-between', minWidth: '220px', maxWidth: '420px', whiteSpace: 'nowrap' }}
                  text={(
                  selectedGithubRepo
                    ? <>{selectedGithubRepo.title} <span style={{ fontSize: '10px', color: Colors.GRAY3 }}>[{selectedGithubRepo.value}]</span></>
                    : 'Select Repository')}
                  rightIcon='double-caret-vertical'
                  fill
                />
              </Select>
              <Button
                icon='eraser'
                intent={Intent.WARNING}
                disabled={isRunning || !isEnabled(providerId)}
                onClick={() => setSelectedGithubRepo(null)}
              />
            </ButtonGroup>
          </FormGroup>
        </>
      )
      break
    case Providers.REFDIFF:
      providerSettings = (
        <RefDiffSettings
          providerId={providerId}
          repoId={refDiffRepoId}
          tasks={refDiffTasks}
          pairs={refDiffPairs}
          setRepoId={setRefDiffRepoId}
          setTasks={setRefDiffTasks}
          setPairs={setRefDiffPairs}
          isRunning={isRunning}
          isEnabled={isEnabled}
        />
      )
      break
    default:
      providerSettings = null
      break
  }

  return providerSettings
}

export default ProviderSettings
