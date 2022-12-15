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
import { useLocation, useHistory } from 'react-router-dom'
import {
  Menu,
  MenuItem,
  Navbar,
  Icon,
  Button,
  Alignment,
  Position,
  Intent
} from '@blueprintjs/core'
import { Popover2 } from '@blueprintjs/popover2'

import { Logo } from '@/components'

import SlackIcon from '@/images/slack-mark-monochrome-black.svg'
import SlackLogo from '@/images/slack-rgb.svg'

import { useMenu, MenuItemType } from './use-menu'
import * as S from './styled'
import {useVersion} from "@/store";

interface Props {
  children: React.ReactNode
}

export const BaseLayout = ({ children }: Props) => {
  const menu = useMenu()
  const { pathname } = useLocation()
  const history = useHistory()
  const { version } = useVersion()

  const handlePushPath = (it: MenuItemType) => {
    if (!it.target) {
      history.push(it.path)
    } else {
      window.open(it.path, '_blank')
    }
  }

  return (
    <S.Container>
      <S.Sider>
        <Logo />
        <Menu className='menu'>
          {menu.map((it) => (
            <MenuItem
              key={it.key}
              className='menu-item'
              text={it.title}
              icon={it.icon}
              active={pathname.includes(it.path)}
              onClick={() => handlePushPath(it)}
            >
              {it.children?.map((cit) => (
                <MenuItem
                  key={cit.key}
                  className='sub-menu-item'
                  text={cit.title}
                  icon={cit.icon ?? <img src={cit.iconUrl} width={16} alt='' />}
                  active={pathname.includes(cit.path)}
                  onClick={() => handlePushPath(cit)}
                />
              ))}
            </MenuItem>
          ))}
        </Menu>
        <div className='copyright'>
          <div>{version}</div>
          <div>Apache 2.0 License</div>
        </div>
      </S.Sider>
      <S.Inner>
        <S.Header>
          <Navbar.Group align={Alignment.RIGHT}>
            <a
              href='https://github.com/apache/incubator-devlake'
              rel='noreferrer'
              target='_blank'
            >
              <Icon icon='git-branch' size={16} />
            </a>
            <Navbar.Divider />
            <a
              href='mailto:hello@merico.dev'
              rel='noreferrer'
              target='_blank'
              className='navIconLink'
            >
              <Icon icon='envelope' size={16} />
            </a>
            <Navbar.Divider />
            <Popover2
              position={Position.LEFT}
              content={
                <S.SlackContainer>
                  <img src={SlackLogo} alt='' />
                  <p>
                    <span>Want to interact with the </span>
                    <strong>Merico Community</strong>
                    <span>? Join us on our Slack Channel.</span>
                  </p>
                  <p>
                    <a
                      href='https://join.slack.com/t/devlake-io/shared_invite/zt-17b6vuvps-x98pqseoUagM7EAmKC82xQ'
                      rel='noreferrer'
                      target='_blank'
                    >
                      <Button intent={Intent.WARNING}>
                        <span>Message us on </span>
                        <strong>Slack</strong>
                      </Button>
                    </a>
                  </p>
                </S.SlackContainer>
              }
            >
              <img
                src={SlackIcon}
                width={30}
                alt=''
                style={{ cursor: 'pointer' }}
              />
            </Popover2>
          </Navbar.Group>
        </S.Header>
        <S.Content>{children}</S.Content>
      </S.Inner>
    </S.Container>
  )
}
