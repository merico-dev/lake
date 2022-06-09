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

// import '@blueprintjs/popover2/lib/css/blueprint-popover2.css'

export default function GitlabSettings (props) {
  const { connection, provider, onSettingsChange = () => {} } = props
  const history = useHistory()
  // const { providerId, connectionId } = useParams()
  const [jiraBoardGitlabProjects, setJiraBoardGitlabProjects] = useState('')
  // const [showBoardMappingDialog, setShowBoardMappingDialog] = useState(false)
  //
  // const [selectedBoard, setSelectedBoard] = useState({ id: 1, title: 'Open', value: 1 })
  // const [boards, setBoards] = useState(boardsData)
  //
  // const [selectedProject, setSelectedProject] = useState({ id: 0, title: 'GL PRJ 3E4', value: 938191 })
  // const [projects, setProjects] = useState(projectsData)
  //
  // const [boardMappings, setBoardMappings] = useState([
  //   {
  //     id: 0,
  //     boards: [8],
  //     projects: [8967944, 8967945],
  //     deleted: false
  //   },
  //   {
  //     id: 1,
  //     boards: [9],
  //     projects: [8967946, 8967947],
  //     deleted: false
  //   },
  //   {
  //     id: 2,
  //     boards: [10],
  //     projects: [8967946, 8967947, 1967900],
  //     deleted: false
  //   }
  // ])
  // const [deletedBoards, setDeletedBoards] = useState([])
  //
  // const boardMapExists = (boardMap) => {
  //   return boardMappings.some(b => b.boards.includes(boardMap.boards[0]) && b.projects.includes(boardMap.projects[0]))
  // }

  useEffect(() => {
    const settings = {
      JIRA_BOARD_GITLAB_PROJECTS: jiraBoardGitlabProjects
    }
    onSettingsChange(settings)
    console.log('>> GITLAB INSTANCE SETTINGS FIELDS CHANGED!', settings)
  }, [
    jiraBoardGitlabProjects,
    onSettingsChange
  ])

  // const createBoardMapping = () => {
  //   setShowBoardMappingDialog(true)
  // }
  //
  // const addBoardMapping = (boardMap) => {
  //   setBoardMappings([
  //     ...boardMappings,
  //     boardMap
  //   ])
  // }
  //
  // const editBoardMapping = () => {
  //
  // }
  //
  // const deleteBoardMapping = (boardMap) => {
  //   const newBoards = boardMappings.filter(board => board.id !== boardMap.id)
  //   setBoardMappings(newBoards)
  //   console.log('>>> EDIT BOARD MAPPINGS', newBoards)
  // }
  //
  // const linkProject = () => {
  //   setShowBoardMappingDialog(false)
  //   // @todo MULTI-PROJECT MAPPING!
  //   const newBoardMap = {
  //     id: boardMappings.length,
  //     boards: [selectedBoard.value],
  //     projects: [selectedProject.value],
  //     deleted: false
  //   }
  //   if (!boardMapExists(newBoardMap)) {
  //     addBoardMapping(newBoardMap)
  //   } else {
  //     ToastNotification.show({ message: 'Gitlab Project Link already ', intent: 'danger', icon: 'error' })
  //   }
  //   console.log('>> LINKING JIRA BOARD', selectedBoard, ' TO GITLAB PROJECT ', selectedProject)
  // }

  const cancel = () => {
    history.push(`/integrations/${provider.id}`)
  }

  useEffect(() => {
    // @todo FETCH BOARDS FROM BE API
    // setBoards([])

    // @todo FETCH PROJECTS FROM BE API
    // setProjects([])
    if (connection && connection.ID) {
      console.log('>> GITLAB CONNECTION OBJECT RECEIVED...', connection, connection.JIRA_BOARD_GITLAB_PROJECTS)
      setJiraBoardGitlabProjects(connection.JIRA_BOARD_GITLAB_PROJECTS)
    } else {
      console.log('>>>> WARNING!! NO CONNECTION OBJECT', connection)
    }
  }, [connection])

  return (
    <>
      <div className='headlineContainer'>
        <h5>No Additional Settings</h5>
        <p className='description'>
          This integration doesn’t require any configuration.
          You can continue to&nbsp;
          <a href='#' style={{ textDecoration: 'underline' }} onClick={cancel}>add other data connections</a>&nbsp;
          or trigger collection at the <a href='#' style={{ textDecoration: 'underline' }} onClick={cancel}>previous page</a>.
        </p>
      </div>
      {/* #1245 - DISABLE Gitlab Project ID Mappings <DEPRECATED> */}
      {/* <h3 className='headline'>Enter Map IDs Manually</h3>
      <p className=''>Type comma separated mappings using the format <code>[JIRA_BOARD_ID]:[GITLAB_PROJECT_ID]</code></p>
      <div className='formContainer'>
        <FormGroup
          disabled={isSaving}
          labelFor='jira-board-projects'
          helperText='JIRA_BOARD_GITLAB_PROJECTS'
          className='formGroup'
          contentClassName='formGroupContent'
        >
          <Label>
            Map JIRA Boards to <strong>GitLab</strong>
          </Label>
          <InputGroup
            id='jira-storypoint-field'
            placeholder='<JIRA_BOARD>:<GITLAB_PROJECT_ID>,...; eg. 8:8967944,8967945;9:8967946,8967947'
            defaultValue={jiraBoardGitlabProjects}
            onChange={(e) => setJiraBoardGitlabProjects(e.target.value)}
            disabled={isSaving}
            className='input'
          />
        </FormGroup>
      </div> */}
      {/* ========== BOARD LINKING UX ======================== */}
      {/* @todo continue/restore  board-linking ux after ITER3 */}
      {/* <h3 className='headline'>JIRA Board Mappings <span className='bp3-form-helper-text'>JIRA_BOARD_GITLAB_PROJECTS</span></h3>
      <p className=''>Visually relate specific JIRA boards to GitLab Projects.</p> */}

      {/* <Card interactive={false} elevation={Elevation.ONE} style={{ width: '100%', maxWidth: '640px', backgroundColor: Colors.LIGHT_GRAY5 }}>
        {boardMappings.length === 0 && (
          <>
            <h3 style={{ margin: 0 }}>NO BOARD MAPPINGS</h3>
            <p style={{ margin: '0 0 10px 0' }}>No active mappings have been created as yet.</p>
          </>
        )}
        <div className='formContainer'>
          <Button
            className='btn-save'
            icon='add'
            text='Map Gitlab Project'
            loading={isSaving}
            disabled={isSaving}
            onClick={createBoardMapping}
          />
        </div>
        <div className='board-mappings-list' style={{ display: 'flex', flexDirection: 'column' }}>
          {boardMappings.map((boardMap, idx) => (
            <div
              key={`board-entry-key-${idx}`}
              className='board-mapping-entry'
              style={{ display: 'flex', justifyContent: 'space-between', width: '100%', marginBottom: '8px' }}
            >
              <div className='board-mapping' style={{ display: 'flex', width: '100%', minWidth: '240px' }}>
                <div className='board-mapping-from' style={{ marginRight: '10px' }}>
                  {boardMap.boards.map((boardId) => (
                    <Tag
                      key={`board-tag-key${boardId}`}
                      className='node-id-tag board-tag tag-from' intent='success' round fill={false}
                    >BRD {boardId}
                    </Tag>
                  ))}
                </div>
                <div
                  className='board-mapping-linkage'
                  style={{
                    display: 'flex',
                    justifyContent: 'center',
                    flex: 1,
                    margin: 'auto 0',
                    height: '3px',
                    backgroundColor: '#aaaaaa',
                    textAlign: 'center',
                    borderRadius: '6px'
                  }}
                >
                  <Icon icon='symbol-circle' size={12} color={Colors.GRAY2} style={{ alignSelf: 'center', marginRight: 'auto' }} />
                  <Icon
                    icon='flow-review' size={14} color={Colors.GRAY1}
                    style={{ alignSelf: 'center', backgroundColor: '#ffffff', borderRadius: '10px' }}
                  />
                  <Icon icon='symbol-circle' size={12} color={Colors.GRAY2} style={{ alignSelf: 'center', marginLeft: 'auto' }} />
                </div>
                <div className='board-mapping-to' style={{ marginLeft: '10px', flexWrap: 'wrap' }}>
                  {boardMap.projects.map((projectId) => (
                    <Tag
                      key={`project-tag-key${projectId}`}
                      className='node-id-tag project-tag tag-to' intent='danger' round fill={false}
                    >PRJ {projectId}
                    </Tag>
                  ))}
                </div>
              </div>
              <div className='board-mapping-actions' style={{ marginLeft: 'auto', paddingLeft: '15px', whiteSpace: 'nowrap' }}>
                <Button minimal small disabled={isSaving} intent={Intent.DANGER} onClick={() => deleteBoardMapping(boardMap)}>
                  <Icon
                    icon='trash' size={14} color={Colors.GRAY3} style={{ cursor: 'pointer' }}
                  />
                </Button>
                <Button minimal small disabled={isSaving} onClick={editBoardMapping(boardMap)}>
                  <Icon
                    icon='edit' size={14} color={Colors.GRAY3} style={{ cursor: 'pointer' }}
                  />
                </Button>
              </div>
            </div>
          ))}
        </div>
      </Card> */}

      {/* @todo continue/restore board-linking dialog ux after ITER3 */}
      {/* <Dialog
        icon='flows'
        onClose={() => setShowBoardMappingDialog(false)}
        title='Map GitLab Project'
        isOpen={showBoardMappingDialog}
        // onOpened={() => setCustomStatusName('')}
        autoFocus={false}
        className='board-mapping-dialog'
      >
        <div className={Classes.DIALOG_BODY}>
          <div style={{ padding: '20px' }}>
            <h3 style={{ margin: 0 }}>SELECT PROJECT IDs</h3>
            <p>Link a GitLab project to JIRA board(s)</p>
            <div style={{ display: 'flex', width: '100%' }}>
              <div style={{ width: '50%' }}>
                <Select
                  popoverProps={{ usePortal: false }}
                  className='selector-board-id'
                  inline={true}
                  fill={true}
                  items={boards}
                  activeItem={selectedBoard}
                  itemPredicate={(query, item) => item.title.toLowerCase().indexOf(query.toLowerCase()) >= 0}
                  itemRenderer={(item, { handleClick, modifiers }) => (
                    <MenuItem
                      active={modifiers.active}
                      key={item.value}
                      label={item.value}
                      onClick={handleClick}
                      text={item.title}
                    />
                  )}
                  noResults={<MenuItem disabled={true} text='No results.' />}
                  onItemSelect={(item) => {
                    // @todo SET/VERIFY BOARD ID
                    setSelectedBoard(item)
                  }}
                >
                  <Button
                    intent={Intent.PRIMARY}
                    style={{ maxWidth: '260px' }}
                    text={selectedBoard ? `${selectedBoard.title}` : boards[0].title}
                    rightIcon='double-caret-vertical'
                    fill
                  />
                </Select>

              </div>
              <div style={{ width: '16px', alignSelf: 'center' }}>
                <Icon
                  icon='flow-review' size={16} color={Colors.GRAY2}
                  style={{ alignSelf: 'center' }}
                />
              </div>
              <div style={{ width: '50%' }}>
                <Select
                  className='selector-project-id'
                  multiple
                  inline={true}
                  fill={true}
                  items={projects}
                  activeItem={selectedProject}
                  itemPredicate={(query, item) => item.title.toLowerCase().indexOf(query.toLowerCase()) >= 0}
                  itemRenderer={(item, { handleClick, modifiers }) => (
                    <MenuItem
                      active={modifiers.active}
                      key={item.value}
                      label={item.value}
                      onClick={handleClick}
                      text={item.title}
                    />
                  )}
                  noResults={<MenuItem disabled={true} text='No results.' />}
                  onItemSelect={(item) => {
                    // @todo SET/VERIFY PROJECT ID
                    setSelectedProject(item)
                  }}
                >
                  <Button
                    intent={Intent.DANGER}
                    style={{ maxWidth: '260px' }}
                    text={selectedProject ? `${selectedProject.title}` : projects[0].title}
                    rightIcon='double-caret-vertical'
                    fill
                  />
                </Select>
              </div>
            </div>
          </div>
        </div>

        <div className={Classes.DIALOG_FOOTER}>
          <div className={Classes.DIALOG_FOOTER_ACTIONS}>
            <Button
              icon='cross'
              // outlined
              text='Cancel'
              loading={isSaving}
              disabled={isSaving}
              onClick={() => setShowBoardMappingDialog(false)}
            />
            <Button
              icon='flows' intent='primary' text='Link Project'
              loading={isSaving}
              disabled={isSaving}
              onClick={linkProject}
            />
          </div>
        </div>

      </Dialog> */}
    </>
  )
}
