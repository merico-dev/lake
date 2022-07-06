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
import React, { Fragment, useEffect, useState, useCallback } from 'react'
import { CSSTransition } from 'react-transition-group'
import {
  Button,
  Icon,
  Intent,
  Elevation,
  Card,
  Colors,
} from '@blueprintjs/core'

const StandardStackedList = (props) => {
  const {
    items = [],
    transformations = {},
    className = 'selected-items-list',
    connection,
    activeItem,
    editButtonText = 'Edit Transformation',
    addButtonText = 'Add Transformation',
    onAdd = () => {},
    onChange = () => {},
    style = { padding: 0, marginTop: '10px' }
  } = props

  useEffect(() => {
    console.log('>>> Selector List Transformations...', transformations, activeItem)
  }, [transformations, activeItem])

  return (
    <>
      {items[connection.id]?.length > 0 && (
        <Card
          className={className}
          elevation={Elevation.ZERO}
          style={style}
        >
          {items[connection.id]?.map((item, pIdx) => (
            <div
              className='item-entry'
              key={`item-row-key-${pIdx}`}
              style={{
                display: 'flex',
                width: '100%',
                height: '32px',
                lineHeight: '100%',
                justifyContent: 'space-between',
                // margin: '8px 0',
                padding: '8px 12px',
                borderBottom: '1px solid #f0f0f0',
                backgroundColor:
                  activeItem === item
                    ? 'rgba(116, 151, 247, 0.2)'
                    : '#fff',
              }}
            >
              <div>
                <div className='item-name' style={{ fontWeight: 600 }}>
                  {/* <input
                    type='radio'
                    name='configured-item'
                    checked={item === activeItem}
                    onChange={() => onChange(item)}
                  />{' '} */}
                  <label onClick={() => onAdd(item)} style={{ cursor: 'pointer' }}>{item}</label>
                </div>
              </div>
              <div
                style={{
                  display: 'flex',
                  alignContent: 'center',
                }}
              >
                <div
                  className='item-actions'
                  style={{ paddingLeft: '20px' }}
                >
                  <Button
                    intent={Intent.PRIMARY}
                    className='item-action-transformation'
                    icon={
                      <Icon
                        // icon='plus'
                        size={12}
                        color={Colors.BLUE4}
                      />
                    }
                    text={Object.keys(transformations[item] ? transformations[item] : {})?.some(configObject => transformations[item][configObject]?.length > 0 )  ? editButtonText : addButtonText }
                    color={Colors.BLUE3}
                    small
                    minimal={activeItem !== item ? true : false}
                    style={{
                      minWidth: '18px',
                      minHeight: '18px',
                      fontSize: '11px',
                    }}
                    onClick={() => onAdd(item)}
                  />
                </div>
              </div>
            </div>
          ))}
        </Card>
      )}
    </>
  )
}

export default StandardStackedList
