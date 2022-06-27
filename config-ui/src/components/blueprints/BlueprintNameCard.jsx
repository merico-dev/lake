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
import { InputGroup, Divider, Elevation, Card } from '@blueprintjs/core'
import InputValidationError from '@/components/validation/InputValidationError'

const BlueprintNameCard = (props) => {
  const {
    activeStep,
    name,
    setBlueprintName = () => {},
    fieldHasError = () => {},
    getFieldError = () => {},
    advancedMode = false,
  } = props

  return (
    <Card
      className="workflow-card"
      elevation={Elevation.TWO}
      style={{ width: '100%' }}
    >
      <h3>
        {advancedMode ? 'Advanced' : ''} Blueprint Name{' '}
        <span className="required-star">*</span>
      </h3>
      <Divider className="section-divider" />
      <p>
        Give your Blueprint a unique name to help you identify it in the future.
      </p>
      <InputGroup
        id="blueprint-name"
        placeholder="Enter Blueprint Name"
        value={name}
        onChange={(e) => setBlueprintName(e.target.value)}
        className={`blueprint-name-input ${
          fieldHasError('Blueprint Name') ? 'invalid-field' : ''
        }`}
        inline={true}
        style={{ marginBottom: '10px' }}
        rightElement={
          <InputValidationError error={getFieldError('Blueprint Name')} />
        }
      />
    </Card>
  )
}

export default BlueprintNameCard
