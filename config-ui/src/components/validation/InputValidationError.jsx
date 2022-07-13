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
import React, { useEffect, useState, useCallback } from 'react'
import {
  Colors,
  Icon,
  Popover,
  PopoverInteractionKind,
  Intent,
  Position,
} from '@blueprintjs/core'

const InputValidationError = (props) => {
  const {
    error,
    position = Position.TOP,
    // eslint-disable-next-line no-unused-vars
    validateOnFocus = false,
    elementRef,
    onError = () => {},
    onSuccess = () => {},
    interactionKind = PopoverInteractionKind.HOVER_TARGET_ONLY,
  } = props

  const [elementIsFocused, setElementIsFocused] = useState(false)
  // eslint-disable-next-line no-unused-vars
  const [inputElement, setInputElement] = useState(null)

  const handleElementFocus = useCallback(
    (isFocused, ref) => {
      setElementIsFocused(isFocused)
      if (error) {
        elementRef?.current.parentElement.classList.remove('valid-field')
        elementRef?.current.parentElement.classList.add('invalid-field')
      } else {
        elementRef?.current.parentElement.classList.remove('invalid-field')
        elementRef?.current.parentElement.classList.add('valid-field')
      }
    },
    [elementRef, error]
  )

  const handleElementBlur = useCallback(
    (isFocused, ref) => {
      setElementIsFocused(isFocused)
      if (!error) {
        elementRef?.current.parentElement.classList.remove('invalid-field')
      }
    },
    [elementRef, error]
  )

  useEffect(() => {
    const iRef = elementRef?.current
    if (iRef) {
      setInputElement(iRef)
      iRef.addEventListener(
        'focus',
        (e) => handleElementFocus(true, iRef),
        true
      )
      iRef.addEventListener(
        'keyup',
        (e) => handleElementFocus(true, iRef),
        true
      )
      iRef.addEventListener('blur', (e) => handleElementBlur(false, iRef), true)
    } else {
      setInputElement(null)
    }

    return () => {
      iRef?.removeEventListener('focus', setElementIsFocused, true)
      iRef?.removeEventListener('keyup', setElementIsFocused, true)
      iRef?.removeEventListener('blur', setElementIsFocused, true)
      setInputElement(null)
    }
  }, [elementRef, handleElementBlur, handleElementFocus])

  useEffect(() => {
    if (error && validateOnFocus && elementIsFocused) {
      onError(elementRef?.current?.id ? elementRef?.current?.id : null)
    } else if (error && !validateOnFocus) {
      onError(elementRef?.current?.id ? elementRef?.current?.id : null)
    } else {
      onSuccess()
    }
  }, [error, onError, onSuccess, elementIsFocused, validateOnFocus, elementRef])

  useEffect(() => {}, [validateOnFocus])

  return error ? (
    <div
      className="inline-input-error"
      style={{ outline: 'none', cursor: 'pointer', margin: '5px 5px 3px 5px' }}
    >
      <Popover
        position={position}
        usePortal={true}
        openOnTargetFocus={true}
        intent={Intent.WARNING}
        interactionKind={interactionKind}
        enforceFocus={false}
        // autoFocus={false}
      >
        <Icon
          icon="warning-sign"
          size={12}
          color={
            (validateOnFocus && elementIsFocused) || (error && !validateOnFocus)
              ? Colors.RED5
              : Colors.GRAY5
          }
          style={{ outline: 'none' }}
          onClick={(e) => e.stopPropagation()}
        />
        <div
          style={{
            outline: 'none',
            padding: '5px',
            borderTop: `2px solid ${Colors.RED5}`,
          }}
        >
          {error}
        </div>
      </Popover>
    </div>
  ) : null
}

export default InputValidationError
