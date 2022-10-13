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
import React, { useState, useEffect, useMemo } from 'react'
import {
  Intent,
  FormGroup,
  RadioGroup,
  InputGroup,
  Radio,
  Tag
} from '@blueprintjs/core'
import { Providers, ProviderLabels } from '@/data/Providers'

const Deployment = (props) => {
  const {
    provider,
    transformation,
    entityIdKey,
    isSaving = false,
    onSettingsChange = () => {}
  } = props

  const [selectValue, setSelectValue] = useState(1)

  useEffect(() => {
    setSelectValue(
      transformation?.deploymentPattern ||
        transformation?.deploymentPattern === '' ||
        transformation?.productionPattern ||
        transformation?.productionPattern === ''
        ? 1
        : 0
    )
  }, [transformation?.deploymentPattern, transformation?.productionPattern])

  const handleChangeSelectValue = (sv) => {
    if (entityIdKey && sv === 0) {
      onSettingsChange(
        { deploymentPattern: undefined, productionPattern: undefined },
        entityIdKey
      )
    } else if (entityIdKey && sv === 1) {
      onSettingsChange(
        { deploymentPattern: '', productionPattern: '' },
        entityIdKey
      )
    }
    setSelectValue(sv)
  }

  const radioLabels = useMemo(() => {
    let radio1
    let radio2

    const providerId = provider?.id
    const providerName = ProviderLabels[provider?.id?.toUpperCase()]

    switch (providerId) {
      case Providers.JENKINS:
        radio1 = 'Detect Deployment from Builds in Jenkins'
        radio2 = 'Not using Jenkins Builds as Deployments'
        break
      case Providers.GITHUB:
        radio1 = `Detect Deployment from Jobs in GitHub Action`
        radio2 = 'Not using Jobs in GitHub Action as Deployments'
        break
      case Providers.GITLAB:
      default:
        radio1 = `Detect Deployment from Jobs in ${providerName} CI`
        radio2 = `Not using ${providerName} Builds as Deployments`
    }

    return [radio1, radio2]
  }, [provider])

  const tagHints = useMemo(() => {
    let hint1
    let hint2

    const providerId = provider?.id

    switch (providerId) {
      case Providers.JENKINS:
        hint1 =
          'A Jenkins build with a name that matches the given regEx will be considered as a Deployment.'
        hint2 =
          // eslint-disable-next-line max-len
          'A Jenkins build that matches the given regEx will be considered as a build in the Production environment. If you leave this field empty, all data will be tagged as in the Production environment.'
        break
      case Providers.GITHUB:
        hint1 =
          'A GitHub Action job with a name that matches the given regEx will be considered as a Deployment.'
        hint2 =
          // eslint-disable-next-line max-len
          'A GitHub Action job that matches the given regEx will be considered as a job in the Production environment. If you leave this field empty, all data will be tagged as in the Production environment.'
        break
      case Providers.GITLAB:
        hint1 =
          'A GitLab CI job with a name that matches the given regEx will be considered as a Deployment.'
        hint2 =
          // eslint-disable-next-line max-len
          'A GitLab CI job that matches the given regEx will be considered as a job in the Production environment. If you leave this field empty, all data will be tagged as in the Production environment.'
        break
    }

    return [hint1, hint2]
  }, [provider])

  return (
    <>
      <h5>CI/CD</h5>
      <p style={{ color: '#292B3F' }}>
        <strong>What is a deployment?</strong>{' '}
        <Tag intent={Intent.PRIMARY} style={{ fontSize: '10px' }} minimal>
          DORA
        </Tag>
      </p>
      <p>Define Deployment using one of the following options.</p>
      <RadioGroup
        inline={false}
        label={false}
        name='deploy-tag'
        onChange={(e) => handleChangeSelectValue(+e.target.value)}
        selectedValue={selectValue}
        required
      >
        <Radio label={radioLabels[0]} value={1} />
        {selectValue === 1 && (
          <>
            <p>{tagHints[0]}</p>
            <div className='formContainer'>
              <FormGroup
                disabled={isSaving}
                inline={true}
                label={
                  <label
                    className='bp3-label'
                    style={{ minWidth: '150px', marginRight: '10px' }}
                  >
                    Deployment
                  </label>
                }
                labelFor='deploy-tag-production'
                className='formGroup'
                contentClassName='formGroupContent'
              >
                <InputGroup
                  id='deploy-tag-production'
                  placeholder='(?i)deploy'
                  value={transformation?.deploymentPattern}
                  onChange={(e) =>
                    onSettingsChange(
                      { deploymentPattern: e.target.value },
                      entityIdKey
                    )
                  }
                  disabled={isSaving}
                  className='input'
                  maxLength={255}
                />
              </FormGroup>
            </div>
            <p>{tagHints[1]}</p>
            <FormGroup
              disabled={isSaving}
              inline={true}
              label={
                <label
                  className='bp3-label'
                  style={{ minWidth: '150px', marginRight: '10px' }}
                >
                  Production
                </label>
              }
              labelFor='production'
              className='formGroup'
              contentClassName='formGroupContent'
            >
              <InputGroup
                id='deploy-tag-production'
                placeholder='(?i)production'
                value={transformation?.productionPattern}
                onChange={(e) =>
                  onSettingsChange(
                    { productionPattern: e.target.value },
                    entityIdKey
                  )
                }
                disabled={isSaving}
                className='input'
                maxLength={255}
              />
            </FormGroup>
          </>
        )}
        <Radio label={radioLabels[1]} value={0} />
      </RadioGroup>
    </>
  )
}

export default Deployment
