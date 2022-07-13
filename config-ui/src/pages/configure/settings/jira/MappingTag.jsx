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
import { FormGroup, Label, Tag, TagInput } from '@blueprintjs/core'

const MappingTag = ({ classNames, labelIntent, labelName, onChange, rightElement, helperText, typeOrStatus, values, placeholderText }) => {
  return (
    <>
      <div className='formContainer'>
        <FormGroup
          // disabled={isTesting || isSaving}
          label=''
          inline={true}
          labelFor='jira-issue-type-mapping'
          helperText={helperText}
          className='formGroup'
          contentClassName='formGroupContent'
        >
          {labelName && (
            <Label style={{ display: 'inline' }}>
              <span style={{ marginRight: '10px' }}><Tag className={classNames} intent={labelIntent}>{labelName}</Tag></span>
            </Label>
          )}
          <TagInput
            placeholder={placeholderText}
            values={values || []}
            fill={true}
            onChange={value => setTimeout(() => onChange([...new Set(value)]), 0)}
            addOnPaste={true}
            addOnBlur={true}
            rightElement={rightElement}
            onKeyDown={e => e.key === 'Enter' && e.preventDefault()}
            className='tagInput'
          />
        </FormGroup>
      </div>
    </>
  )
}

export default MappingTag
