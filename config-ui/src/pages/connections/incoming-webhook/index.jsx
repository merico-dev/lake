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

import React, { useState } from 'react'
import { Link } from 'react-router-dom'
import { Icon, Button, Intent } from '@blueprintjs/core'

import Nav from '@/components/Nav'
import Sidebar from '@/components/Sidebar'
import Content from '@/components/Content'
import AppCrumbs from '@/components/Breadcrumbs'
import { useWebhookManager } from '@/hooks/useWebhookManager'
import { ReactComponent as WebHookProviderIcon } from '@/images/integrations/incoming-webhook.svg'
import { ReactComponent as EditIcon } from '@/images/icons/setting-con.svg'
import { ReactComponent as DeleteIcon } from '@/images/icons/delete.svg'

import { AddModal } from './add-modal'
import { ViewOrEditModal } from './view-or-edit-modal'
import { DeleteModal } from './delete-modal'
import * as S from './styled'

const postUrlPrefix = `${window.location.origin}/api`

export const IncomingWebhook = () => {
  // defined the form modal is add | edit | delete
  const [modalType, setModalType] = useState()
  // defined the edit or delete record
  const [record, setRecord] = useState()

  const { loading, data, operating, onCreate, onUpdate, onDelete } =
    useWebhookManager()

  const handleShowModal = (mt, r) => {
    setModalType(mt)
    setRecord((existingRecord) =>
      r
        ? {
            ...r,
            postIssuesEndpoint: `${postUrlPrefix}${r.postIssuesEndpoint}`,
            closeIssuesEndpoint: `${postUrlPrefix}${r.closeIssuesEndpoint}`,
            postDeploymentsCurl: `curl ${postUrlPrefix}${r.postPipelineDeployTaskEndpoint} -X 'POST' -d "{
  \\"commit_sha\\":\\"the sha of deployment commit\\",
  \\"repo_url\\":\\"the repo URL of the deployment commit\\",
  \\"start_time\\":\\"Optional, eg. 2020-01-01T12:00:00+00:00\\"
}"`
          }
        : existingRecord
    )
  }

  const handleHideModal = () => {
    setModalType()
    setRecord()
  }

  return (
    <div className='container'>
      <Nav />
      <Sidebar />
      <Content>
        <div className='main'>
          <AppCrumbs
            items={[
              { href: '/', icon: false, text: 'Dashboard' },
              // use /connections replace here
              { href: '/integrations', icon: false, text: 'Integrations' },
              {
                href: '/connections/incoming-webhook',
                icon: false,
                text: 'Incoming Webhook',
                current: true
              }
            ]}
          />
          <div className='headlineContainer'>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: 12
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <WebHookProviderIcon
                  className='providerIconSvg'
                  width='30'
                  height='30'
                />
                <h1 style={{ margin: '0 0 0 8px' }}>Incoming Webhook</h1>
              </div>
              <Link style={{ color: '#777777' }} to='/integrations'>
                <Icon icon='undo' size={16} /> Go Back
              </Link>
            </div>
            <div className='page-description'>
              Use Webhooks to define Incidents and Deployments for your CI tools
              if they are not listed in Data Sources.
            </div>
          </div>
          <div className='manageProvider'>
            <S.Container>
              <span>
                <Button
                  intent='primary'
                  text='Add Incoming Webhook'
                  loading={operating}
                  onClick={() => handleShowModal('add')}
                />
              </span>
              <S.Wrapper>
                <S.Grid className='title'>
                  <li>ID</li>
                  <li>Incoming Webhook Name</li>
                  <li />
                </S.Grid>
                {loading ? (
                  <div>Loading</div>
                ) : (
                  data.map((it) => (
                    <S.Grid key={it.id}>
                      <li>{it.id}</li>
                      <li>{it.name}</li>
                      <li>
                        <Button
                          loading={operating}
                          intent={Intent.PRIMARY}
                          minimal
                          small
                          icon={<EditIcon width={18} height={18} />}
                          onClick={() => handleShowModal('edit', it)}
                        />
                        <Button
                          loading={operating}
                          intent={Intent.PRIMARY}
                          minimal
                          small
                          icon={<DeleteIcon width={18} height={18} />}
                          onClick={() => handleShowModal('delete', it)}
                        />
                      </li>
                    </S.Grid>
                  ))
                )}
              </S.Wrapper>
            </S.Container>
          </div>
        </div>
      </Content>
      {modalType === 'add' && (
        <AddModal onSubmit={onCreate} onCancel={handleHideModal} />
      )}
      {modalType === 'edit' && (
        <ViewOrEditModal
          record={record}
          onSubmit={onUpdate}
          onCancel={handleHideModal}
        />
      )}
      {modalType === 'delete' && (
        <DeleteModal
          record={record}
          onSubmit={onDelete}
          onCancel={handleHideModal}
        />
      )}
    </div>
  )
}
