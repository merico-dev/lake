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

import styled from '@emotion/styled'

export const Container = styled.div`
  margin-top: 48px;
`

export const Wrapper = styled.div`
  margin-top: 12px;
  background: #ffffff;
  box-shadow: 0px 2.4px 4.8px -0.8px rgba(0, 0, 0, 0.1),
    0px 1.6px 8px rgba(0, 0, 0, 0.07);
  border-radius: 4px;
`

export const Grid = styled.ul`
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  align-items: center;
  padding: 0 16px;
  width: 100%;
  height: 48px;
  font-size: 14px;
  color: #292b3f;
  border-top: 1px solid #bdcefb;
  box-sizing: border-box;

  &.title {
    font-size: 16px;
    font-weight: 600;
  }

  &:first-child {
    border-top: 0;
  }

  li {
    flex: auto;

    &:first-child {
      flex: 0 0 200px;
    }

    &:last-child {
      flex: 0 0 60px;
      display: flex;
      justify-content: space-around;

      & > svg {
        cursor: pointer;
      }
    }
  }
`

export const FormWrapper = styled.div`
  padding: 24px;
  font-size: 14px;
  color: #292b3f;

  h2 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }

  h3 {
    margin: 16px 0 0;
    font-size: 14px;
    font-weight: 600;
  }

  p {
    margin: 8px 0;
    font-size: 12px;
    color: #94959f;
  }

  .message {
    margin-bottom: 24px;

    p {
      font-size: 14px;
    }
  }

  .form {
    input {
      padding: 7px 12px;
      width: 100%;
      height: 32px;
      background-color: #ffffff;
      border: 1px solid #dbe4fd;
      border-radius: 4px;
      outline: none;
      box-sizing: border-box;

      &.has-error {
        border: 1px solid red;
      }
    }

    .error {
      color: red;
    }
  }

  .tips {
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 16px;

    span {
      margin-left: 4px;
    }
  }

  .url {
    margin-top: 24px;

    .block {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 10px;
      background-color: #f0f4fe;

      & > span {
        flex: 1 0;
      }

      & > svg {
        cursor: pointer;
      }
    }
  }

  .btns {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    margin-top: 24px;

    .bp3-button + .bp3-button {
      margin-left: 8px;
    }
  }
`
