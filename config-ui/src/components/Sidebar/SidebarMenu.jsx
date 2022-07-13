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
import React, { useEffect, Fragment } from 'react'
import { Menu } from '@blueprintjs/core'
import '@/styles/sidebar-menu.scss'

const SidebarMenu = (props) => {
  const { menu } = props
  // const activeRoute = useRouteMatch()

  useEffect(() => {

  }, [menu])

  return (
    <>
      <Menu className='sidebarMenu' style={{ marginTop: '10px' }}>
        {menu.map((m, mIdx) => (
          m.children.length === 0
            ? (
              <Menu.Item
                active={m.active}
                key={`menu-item-key${mIdx}`}
                icon={m.icon}
                text={m.label}
                href={m.route}
                target={m.target}
                disabled={m.disabled}
              />
              )
            : (
              <Menu.Item
                className='is-submenu has-children'
                active={m.active}
                key={`menu-item-key${mIdx}`}
                text={m.label}
                icon={m.icon}
                href={m.route}
                disabled={m.disabled}
              >
                {m.children.map((mS, mSidx) => (
                  <Menu.Item
                    active={mS.active}
                    key={`submenu-${mIdx}-item-key${mSidx}`}
                    href={mS.route}
                    icon={mS.icon}
                    text={mS.label}
                    disabled={mS.disabled}
                    // className={mS.classNames.join(' ')}
                  />))}
              </Menu.Item>
              )
        ))}
        <Menu.Divider />
      </Menu>
    </>
  )
}

export default SidebarMenu
