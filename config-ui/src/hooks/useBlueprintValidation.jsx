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
import { useCallback, useEffect, useState } from 'react'
import parser from 'cron-parser'
import { BlueprintMode } from '@/data/NullBlueprint'

function useBlueprintValidation ({
  name,
  cronConfig,
  customCronConfig,
  enable,
  tasks = [],
  mode = null,
  activeStep = null
}) {
  const [errors, setErrors] = useState([])
  const [isValid, setIsValid] = useState(false)

  const clear = () => {
    setErrors([])
  }

  const isValidCronExpression = (expression) => {
    let isValid = false
    try {
      parser.parseExpression(expression)
      isValid = true
    } catch (e) {
      isValid = false
    }
    return isValid
  }

  const validate = useCallback(() => {
    const errs = []
    // console.log('>> VALIDATING BLUEPRINT ', name)

    if (!name) {
      errs.push('Blueprint Name: Enter a valid Name')
    }

    if (name && name.length <= 2) {
      errs.push('Blueprint Name: Name too short, 3 chars minimum.')
    }

    if (!cronConfig) {
      errs.push('Blueprint Cron: No Crontab schedule defined.')
    }

    if (cronConfig && cronConfig !== 'custom' && !isValidCronExpression(cronConfig)) {
      errs.push('Blueprint Cron: Invalid Crontab Expression, unable to parse.')
    }

    if (cronConfig === 'custom' && !isValidCronExpression(customCronConfig)) {
      errs.push(`Blueprint Cron: Invalid Custom Expression, unable to parse. [${customCronConfig}]`)
    }

    if (enable && tasks?.length === 0) {
      errs.push('Blueprint Tasks: Invalid/Empty Configuration')
    }
    
    if (mode !== null && ![BlueprintMode.NORMAL, BlueprintMode.ADVANCED].includes(mode)) {
      errs.push('Invalid / Unsupported Blueprint Mode Detected!')
    }

    setErrors(errs)
  }, [
    name,
    cronConfig,
    customCronConfig,
    tasks,
    enable,
    mode
  ])

  const fieldHasError = useCallback((fieldId) => {
    return errors.some(e => e.includes(fieldId))
  }, [errors])

  const getFieldError = useCallback((fieldId) => {
    return errors.find(e => e.includes(fieldId))
  }, [errors])

  useEffect(() => {
    // console.log('>>> BLUEPRINT FORM ERRORS...', errors)
    setIsValid(errors.length === 0)
    if (errors.length > 0) {
      // ToastNotification.clear()
    }
  }, [errors])

  return {
    errors,
    setErrors,
    isValid,
    validate,
    clear,
    fieldHasError,
    getFieldError
  }
}

export default useBlueprintValidation
