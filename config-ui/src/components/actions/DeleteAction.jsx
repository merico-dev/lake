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
import {
  Button, Colors,
  Position,
  Icon,
  Intent,
  Popover,
  Classes
} from '@blueprintjs/core'

const DeleteAction = (props) => {
  const {
    id, connection, showConfirmation = () => {}, onConfirm = () => {}, onCancel = () => {},
    isDisabled = false,
    isLoading = false,
    text = 'Delete',
    children
  } = props
  return (
    <Popover
      key={`delete-popover-key-${connection.ID}`}
      className='trigger-delete-connection'
      popoverClassName='popover-delete-connection'
      position={Position.RIGHT}
      autoFocus={false}
      enforceFocus={false}
      isOpen={id === connection.ID}
      usePortal={false}
    >
      <a
        href='#'
        intent={Intent.DANGER}
        data-provider={connection.id}
        className='table-action-link actions-link'
        onClick={showConfirmation}
        style={{ color: '#DB3737' }}
      >
        <Icon icon='trash' color={Colors.RED3} size={12} />
        Delete
      </a>
      <>
        <div style={{ padding: '15px 20px 15px 15px' }}>
          {children}
          <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 15 }}>
            <Button
              className={Classes.POPOVER2_DISMISS}
              style={{ marginRight: 10 }}
              disabled={isDisabled || isLoading}
              onClick={onCancel}
            >
              Cancel
            </Button>
            <Button
              disabled={isDisabled}
              loading={isLoading}
              onClick={(e) => onConfirm(connection, e)}
              intent={Intent.DANGER}
              icon='remove'
              className={Classes.POPOVER2_DISMISS}
              style={{ fontWeight: 'bold' }}
            >
              {text}
            </Button>
          </div>
        </div>
      </>
    </Popover>
  )
}

export default DeleteAction
