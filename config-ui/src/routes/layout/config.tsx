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

import {
  AppstoreOutlined,
  ProjectOutlined,
  ExperimentOutlined,
  KeyOutlined,
  DashboardOutlined,
  FileSearchOutlined,
  ApiOutlined,
  GithubOutlined,
  SlackOutlined,
} from '@ant-design/icons';

import { DOC_URL } from '@/release';

type MenuItem = {
  key: string;
  label: string;
  icon?: React.ReactNode;
  children?: MenuItem[];
};

export const menuItems: MenuItem[] = [
  {
    key: '/connections',
    label: 'Connections',
    icon: <AppstoreOutlined rev={undefined} />,
  },
  {
    key: '/projects',
    label: 'Projects',
    icon: <ProjectOutlined rev={undefined} />,
  },
  {
    key: '/advanced',
    label: 'Advanced',
    icon: <ExperimentOutlined rev={undefined} />,
    children: [
      {
        key: '/advanced/blueprints',
        label: 'Blueprints',
      },
      {
        key: '/advanced/pipelines',
        label: 'Pipelines',
      },
    ],
  },
  {
    key: '/keys',
    label: 'API Keys',
    icon: <KeyOutlined rev={undefined} />,
  },
];

const getMenuMatchs = (items: MenuItem[], parentKey?: string) => {
  return items.reduce((pre, cur) => {
    pre[cur.key] = {
      ...cur,
      parentKey,
    };

    if (cur.children) {
      pre = { ...pre, ...getMenuMatchs(cur.children, cur.key) };
    }

    return pre;
  }, {} as Record<string, MenuItem & { parentKey?: string }>);
};

export const menuItemsMatch = getMenuMatchs(menuItems);

export const headerItems = [
  {
    link: import.meta.env.DEV ? `${window.location.protocol}//${window.location.hostname}:3002` : `/grafana`,
    label: 'Dashboards',
    icon: <DashboardOutlined rev={undefined} />,
  },
  {
    link: DOC_URL.TUTORIAL,
    label: 'Docs',
    icon: <FileSearchOutlined rev={undefined} />,
  },
  {
    link: '/api/swagger/index.html',
    label: 'API',
    icon: <ApiOutlined rev={undefined} />,
  },
  {
    link: 'https://github.com/apache/incubator-devlake',
    label: 'GitHub',
    icon: <GithubOutlined rev={undefined} />,
  },
  {
    link: 'https://join.slack.com/t/devlake-io/shared_invite/zt-26ulybksw-IDrJYuqY1FrdjlMMJhs53Q',
    label: 'Slack',
    icon: <SlackOutlined rev={undefined} />,
  },
];
