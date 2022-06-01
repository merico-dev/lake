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
import { useEffect, useState, useCallback } from 'react'
import request from '@/utils/request'
import { ToastNotification } from '@/components/Toast'

const useJIRA = ({ apiProxyPath, issuesEndpoint, fieldsEndpoint, boardsEndpoint }) => {
  const [isFetching, setIsFetching] = useState(false)
  const [issueTypes, setIssueTypes] = useState([])
  const [fields, setFields] = useState([])
  const [boards, setBoards] = useState([])
  const [issueTypesResponse, setIssueTypesResponse] = useState([])
  const [fieldsResponse, setFieldsResponse] = useState([])
  const [boardsResponse, setBoardsResponse] = useState([])
  const [error, setError] = useState()

  const fetchIssueTypes = useCallback(() => {
    try {
      if (apiProxyPath.includes('null')) {
        throw new Error('Connection ID is Null')
      }
      setError(null)
      const fetchIssueTypes = async () => {
        const issues = await
        request
          .get(issuesEndpoint)
          .catch(e => setError(e))
        console.log('>>> JIRA API PROXY: Issues Response...', issues)
        setIssueTypesResponse(issues && Array.isArray(issues.data) ? issues.data : [])
        setTimeout(() => {
          setIsFetching(false)
        }, 1000)
      }
      fetchIssueTypes()
    } catch (e) {
      setIsFetching(false)
      setError(e)
      ToastNotification.show({ message: e.message, intent: 'danger', icon: 'error' })
    }
  }, [issuesEndpoint, apiProxyPath])

  const fetchFields = useCallback(() => {
    try {
      if (apiProxyPath.includes('null')) {
        throw new Error('Connection ID is Null')
      }
      setError(null)
      const fetchIssueFields = async () => {
        const fields = await
        request
          .get(fieldsEndpoint)
          .catch(e => setError(e))
        console.log('>>> JIRA API PROXY: Fields Response...', fields)
        setFieldsResponse(fields && Array.isArray(fields.data) ? fields.data : [])
        setTimeout(() => {
          setIsFetching(false)
        }, 1000)
      }
      fetchIssueFields()
    } catch (e) {
      setIsFetching(false)
      setError(e)
      ToastNotification.show({ message: e.message, intent: 'danger', icon: 'error' })
    }
  }, [fieldsEndpoint, apiProxyPath])
  
  const fetchBoards = useCallback(() => {
    try {
      if (apiProxyPath.includes('null')) {
        throw new Error('Connection ID is Null')
      }
      setError(null)
      const fetchBoards = async () => {
        const boards = await
        request
          .get(boardsEndpoint)
          .catch(e => setError(e))
        console.log('>>> JIRA API PROXY: Boards Response...', boards)
        setBoardsResponse(boards && Array.isArray(boards.data) ? boards.data : [])
        setTimeout(() => {
          setIsFetching(false)
        }, 1000)
      }
      fetchBoards()
    } catch (e) {
      setIsFetching(false)
      setError(e)
      ToastNotification.show({ message: e.message, intent: 'danger', icon: 'error' })
    }
  }, [fetchBoards, apiProxyPath])

  const createListData = (data = [], titleProperty = 'name', valueProperty = 'name') => {
    return data.map((d, dIdx) => {
      return {
        ...d,
        id: dIdx,
        key: d.key ? d.key : dIdx,
        title: d[titleProperty],
        value: d[valueProperty],
        type: d.schema?.type || 'string'
      }
    })
  }

  useEffect(() => {
    setIssueTypes(issueTypesResponse
      ? createListData(issueTypesResponse).reduce((pV, cV) => !pV.some(i => i.value === cV.value) ? [...pV, cV] : [...pV], [])
      : [])
  }, [issueTypesResponse])

  useEffect(() => {
    setFields(fieldsResponse ? createListData(fieldsResponse, 'name', 'key') : [])
  }, [fieldsResponse])
  
  useEffect(() => {
    setBoards(boardsResponse ? createListData(boardsResponse, 'name', 'key') : [])
  }, [boardsResponse])

  useEffect(() => {
    console.log('>>> JIRA API PROXY: FIELD SELECTOR LIST DATA', fields)
  }, [fields])

  useEffect(() => {
    console.log('>>> JIRA API PROXY: FIELD SELECTOR BOARDS DATA', boards)
  }, [boards])

  useEffect(() => {
    if (error) {
      console.log('>>> JIRA PROXY API ERROR!', error)
    }
  }, [error])

  return {
    isFetching,
    fetchFields,
    fetchIssueTypes,
    createListData,
    issueTypesResponse,
    fieldsResponse,
    boardsResponse,
    issueTypes,
    fields,
    boards,
    error
  }
}

export default useJIRA
