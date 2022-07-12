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
import axios from 'axios'

const headers = {}

const handleErrorResponse = (e) => {
  let errorResponse = { success: false, message: e.message, data: null, status: 504 }
  if (e.response) {
    errorResponse = { ...errorResponse, data: e.response ? e.response.data : null, message: e.response?.data?.message, status: e.response ? e.response.status : 504 }
  }
  return errorResponse
}

export default {
  post: async (url, body) => {
    return await axios.post(
      url,
      body,
      {
        headers
      }
    ).catch(e => handleErrorResponse(e))
  },
  get: async (url) => {
    return await axios.get(
      url,
      {
        headers
      }
    ).catch(e => handleErrorResponse(e))
  },
  put: async (url, body) => {
    return await axios.put(
      url,
      body,
      {
        headers
      }
    ).catch(e => handleErrorResponse(e))
  },
  patch: async (url, body) => {
    return await axios.patch(
      url,
      body,
      {
        headers
      }
    ).catch(e => handleErrorResponse(e))
  },
  delete: async (url, body) => {
    return await axios.delete(
      url,
      body,
      {
        headers
      }
    ).catch(e => handleErrorResponse(e))
  }
}
