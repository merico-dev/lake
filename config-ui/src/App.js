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

import React from 'react'
import {
  BrowserRouter as Router,
  Route,
} from 'react-router-dom'

import 'normalize.css'
import '@/styles/app.scss'
import 'typeface-montserrat'
import 'jetbrains-mono'
// Theme variables (@styles/theme.scss) injected via Webpack w/ @sass-loader additionalData option!
// import '@/styles/theme.scss'

import Configure from './pages/configure/index'
import Integration from '@/pages/configure/integration/index'
import ManageIntegration from '@/pages/configure/integration/manage'
import AddConnection from '@/pages/configure/connections/AddConnection'
import EditConnection from '@/pages/configure/connections/EditConnection'
import ConfigureConnection from '@/pages/configure/connections/ConfigureConnection'
import Triggers from '@/pages/triggers/index'
import Offline from '@/pages/offline/index'
import Pipelines from '@/pages/pipelines/index'
import CreatePipeline from '@/pages/pipelines/create'
import PipelineActivity from '@/pages/pipelines/activity'
import Blueprints from '@/pages/blueprints/index'
import CreateBlueprint from '@/pages/blueprints/create-blueprint'

function App () {
  return (
    <Router>
      {/* Admin */}
      <Route exact path='/'>
        <Integration />
      </Route>
      <Route path='/integrations/:providerId'>
        <ManageIntegration />
      </Route>
      <Route path='/connections/add/:providerId'>
        <AddConnection />
      </Route>
      <Route path='/connections/edit/:providerId/:connectionId'>
        <EditConnection />
      </Route>
      <Route path='/connections/configure/:providerId/:connectionId'>
        <ConfigureConnection />
      </Route>
      <Route exact path='/integrations'>
        <Integration />
      </Route>
      <Route exact path='/triggers'>
        <Triggers />
      </Route>
      <Route exact path='/pipelines/create'>
        <CreatePipeline />
      </Route>
      <Route exact path='/pipelines'>
        <Pipelines />
      </Route>
      <Route exact path='/pipelines/activity'>
        <PipelineActivity />
      </Route>
      <Route exact path='/pipelines/activity/:pId'>
        <PipelineActivity />
      </Route>
      <Route exact path='/blueprints/create'>
        <CreateBlueprint />
      </Route>
      <Route exact path='/blueprints'>
        <Blueprints />
      </Route>
      <Route exact path='/lake/api/configuration'>
        <Configure />
      </Route>
      <Route exact path='/offline'>
        <Offline />
      </Route>
    </Router>
  )
}

export default App
