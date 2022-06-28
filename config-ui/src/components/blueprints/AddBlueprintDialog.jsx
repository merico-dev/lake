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
import React, { useEffect } from 'react'
import dayjs from '@/utils/time'
import {
  Button,
  ButtonGroup,
  Classes,
  Colors,
  Dialog,
  Elevation,
  FormGroup,
  Icon,
  InputGroup,
  Intent,
  Label,
  MenuItem,
  Popover,
  Position,
  Radio,
  RadioGroup,
  Switch,
  Tooltip,
} from '@blueprintjs/core'
import { Select } from '@blueprintjs/select'
import InputValidationError from '@/components/validation/InputValidationError'
import ContentLoader from '@/components/loaders/ContentLoader'
import PipelineTasks from '@/components/blueprints/PipelineTasks'
import CronHelp from '@/images/cron-help.png'

const AddBlueprintDialog = (props) => {
  const {
    name,
    cronConfig,
    customCronConfig,
    tasks = [],
    enable,
    draftBlueprint,
    isOpen = true,
    isLoading = false,
    selectedPipelineTemplate,
    setIsOpen = () => {},
    setDraftBlueprint = () => {},
    setBlueprintName = () => {},
    setCronConfig = () => {},
    setCustomCronConfig = () => {},
    setBlueprintTasks = () => {},
    setEnableBlueprint = () => {},
    setSelectedPipelineTemplate = () => {},
    // eslint-disable-next-line no-unused-vars
    createCron = () => {},
    // note: nextRunDate helper not operating correctly within Dialog, use createCron instead.
    // eslint-disable-next-line no-unused-vars
    getNextRunDate = (cronExpression) => {},
    saveBlueprint = () => {},
    fieldHasError = () => {},
    getFieldError = () => {},
    getCronPreset = () => {},
    getCronPresetByConfig = () => {},
    isSaving = false,
    isValidBlueprint = false,
    pipelines = [],
    detectedProviders = [],
    tasksLocked = false
  } = props

  useEffect(() => {
  }, [enable, cronConfig, customCronConfig])

  return (
    <>
      <Dialog
        className='dialog-manage-blueprint'
        icon={draftBlueprint ? 'edit' : 'add'}
        title={draftBlueprint ? `Edit ${draftBlueprint.name}` : 'Create Pipeline Blueprint'}
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        onClosed={() => !tasksLocked ? setDraftBlueprint(null) : {}}
        style={{ backgroundColor: '#ffffff' }}
      >
        <div className={Classes.DIALOG_BODY}>
          {isLoading && (
            <ContentLoader
              title='Loading Blueprint...'
              elevation={Elevation.ZERO}
              message='Please wait for configuration.'
            />
          )}
          {!isLoading && (
            <div className='pipeline-form-container'>
              <div className='formContainer' style={{ marginBottom: 0 }}>
                <FormGroup
                  label=''
                  inline={true}
                  labelFor='blueprint-name'
                  className='formGroup-inline'
                  contentClassName='formGroupContent'
                >
                  <Label style={{ display: 'inline', marginRight: 0, fontWeight: 'bold' }}>

                    Blueprint Name

                    <span className='requiredStar'>*</span>
                  </Label>
                  <InputGroup
                    id='blueprint-name'
                    placeholder='Enter Blueprint Name'
                    value={name}
                    onChange={(e) => setBlueprintName(e.target.value)}
                    className={`blueprint-name-input ${fieldHasError('Blueprint Name') ? 'invalid-field' : ''}`}
                    inline={true}
                    style={{ marginBottom: '10px' }}
                    rightElement={(
                      <InputValidationError
                        error={getFieldError('Blueprint Name')}
                      />
                  )}
                  />
                  <Label style={{ display: 'inline', lineHeight: '18px', marginRight: 0, marginBottom: 0, fontWeight: 'bold' }}>
                    Frequency
                    <span className='requiredStar'>*</span>
                  </Label>
                  {getCronPresetByConfig(cronConfig)
                    ? (
                      <small style={{ fontSize: '10px', color: Colors.GRAY2, display: 'block' }}>
                        {getCronPresetByConfig(cronConfig).description}
                      </small>
                      )
                    : ''}
                  <RadioGroup
                    inline={true}
                    label={false}
                    name='blueprint-frequency'
                    onChange={(e) => setCronConfig(e.target.value)}
                    selectedValue={cronConfig}
                    required
                  >
                    {/* Dynamic Presets from Connection Manager */}
                    {[getCronPreset('hourly'),
                      getCronPreset('daily'),
                      getCronPreset('weekly'),
                      getCronPreset('monthly')].map((preset, prIdx) => (
                        <Radio
                          key={`cron-preset-tooltip-key${prIdx}`}
                          label={(
                            <>
                              <Tooltip
                                position={Position.TOP}
                                intent={Intent.PRIMARY}
                                content={preset.description}
                              >{preset.label}
                              </Tooltip>
                            </>
                            )}
                          value={preset.cronConfig}
                          style={{ fontWeight: cronConfig === preset.cronConfig ? 'bold' : 'normal', outline: 'none !important' }}
                        />

                    ))}
                    <Radio label='Custom' value='custom' style={{ fontWeight: cronConfig === 'custom' ? 'bold' : 'normal' }} />
                  </RadioGroup>
                  <div style={{ display: 'flex' }}>
                    <FormGroup
                      disabled={cronConfig !== 'custom'}
                      label=''
                      inline={true}
                      labelFor='cron-custom'
                      className='formGroup-inline'
                      contentClassName='formGroupContent'
                      style={{ marginBottom: '5px' }}
                      fill={false}
                    >
                      <Label style={{ display: 'inline', marginRight: 0, fontWeight: 'bold' }}>
                        Custom Schedule
                        {cronConfig === 'custom' && <span className='requiredStar'>*</span>}
                      </Label>
                      <InputGroup
                        id='cron-custom'
                        readOnly={cronConfig !== 'custom'}
                        leftElement={cronConfig !== 'custom'
                          ? <Icon icon='lock' size={11} style={{ alignSelf: 'center', margin: '4px 10px -2px 6px' }} />
                          : null}
                        rightElement={(
                          <InputValidationError
                            error={getFieldError('Blueprint Cron')}
                          />
                        )}
                        placeholder='Enter Crontab Syntax'
                        value={cronConfig !== 'custom' ? cronConfig : customCronConfig}
                        onChange={(e) => setCustomCronConfig(e.target.value)}
                        className={`cron-custom-input ${fieldHasError('Blueprint Cron') ? 'invalid-field' : ''}`}
                        inline={true}
                        fill={false}
                        style={{ transition: 'none' }}
                      />
                    </FormGroup>
                    <div style={{ marginTop: 'auto', paddingBottom: '15px' }}>
                      <Popover
                        className='trigger-crontab-help'
                        popoverClassName='popover-help-crontab'
                        position={Position.RIGHT}
                        autoFocus={false}
                        enforceFocus={false}
                        usePortal={false}
                      >
                        <a href='#' rel='noreferrer'><Icon icon='help' size={14} style={{ marginLeft: '10px', transition: 'none' }} /></a>
                        <>
                          <div style={{ textShadow: 'none', fontSize: '12px', padding: '12px', maxWidth: '300px' }}>
                            <div style={{
                              marginBottom: '10px',
                              fontWeight: 700,
                              fontSize: '14px',
                            }}
                            >
                              <Icon icon='help' size={16} style={{ marginRight: '5px' }} /> Cron Expression Format
                            </div>
                            <p>Need Help? &mdash; For additional information on <strong>Crontab</strong>,
                              please reference the{' '}
                              <a
                                href='https://man7.org/linux/man-pages/man5/crontab.5.html'
                                rel='noreferrer'
                                target='_blank' style={{ textDecoration: 'underline' }}
                              >
                                Crontab Linux manual
                              </a>.
                            </p>
                            <img src={CronHelp} style={{ border: 0, margin: 0, maxWidth: '100%' }} />
                          </div>
                        </>
                      </Popover>
                    </div>
                  </div>

                </FormGroup>
              </div>
              <div className='formContainer'>
                <FormGroup
                  disabled={isSaving}
                  label={<strong>Pipeline Tasks<span className='requiredStar'>*</span></strong>}
                  labelInfo={tasksLocked
                    ? ''
                    : <span style={{ display: 'block' }}>Choose Pipeline Run Template for task configuration</span>}
                  inline={false}
                  labelFor='blueprint-tasks'
                  className=''
                  contentClassName=''
                  fill
                  style={{ marginBottom: '5px' }}
                >
                  {tasks.length === 0 && (
                    <ButtonGroup>
                      <Select
                        disabled={isSaving || pipelines.length === 0}
                        className='selector-blueprint-tasks'
                        popoverProps={{ usePortal: false, popoverClassName: 'blueprint-tasks-popover', fill: true }}
                        inline={true}
                        fill={true}
                        items={pipelines}
                        activeItem={selectedPipelineTemplate}
                        itemPredicate={(query, item) => item?.title?.toLowerCase().indexOf(query.toLowerCase()) >= 0}
                        itemRenderer={(item, { handleClick, modifiers }) => (
                          <MenuItem
                            active={modifiers.active}
                            key={item.value}
                            label={<small>Run ID #{item.value}</small>}
                            onClick={handleClick}
                            text={
                              <span style={{ paddingRight: '20px' }}>
                                <Icon
                                  color={item.status === 'TASK_COMPLETED' ? Colors.GREEN4 : Colors.GRAY4}
                                  icon={item.status === 'TASK_COMPLETED' ? 'small-tick' : 'dot'} size={11}
                                /> {item.title}
                              </span>
                             }
                          />
                        )}
                        noResults={<MenuItem disabled={true} text='No Pipeline Runs.' />}
                        onItemSelect={(item) => {
                          setSelectedPipelineTemplate(item)
                        }}
                      >
                        <Button
                          className='btn-pipeline-selector'
                          disabled={isSaving}
                          style={{ justifyContent: 'space-between', minWidth: '300px', maxWidth: '460px', whiteSpace: 'nowrap' }}
                          text={selectedPipelineTemplate ? `${selectedPipelineTemplate?.title}` : 'Select Pipeline Run Tasks'}
                          icon='double-caret-vertical'
                          fill={true}
                          rightIcon={(
                            <span style={{ marginRight: '5px' }}>
                              <InputValidationError
                                error={getFieldError('Blueprint Tasks')}
                              />
                            </span>
                      )}
                        />
                      </Select>
                      <Button
                        icon='eraser'
                        intent={Intent.WARNING}
                        disabled={isSaving}
                        onClick={() => setSelectedPipelineTemplate(null) | setBlueprintTasks([])}
                      />
                    </ButtonGroup>
                  )}
                </FormGroup>
              </div>
              <div>
                <PipelineTasks tasks={detectedProviders} />
                {!tasksLocked && tasks.length > 0 &&
                 (
                   <Button
                     onClick={() => setBlueprintTasks([]) | setSelectedPipelineTemplate(null)}
                     icon='eraser'
                     round minimal text='Clear'
                   />
                 )}
              </div>
              <div className='formContainer'>
                <FormGroup
                  label=''
                  inline={true}
                  labelFor='blueprint-enable'
                  className='formGroup-inline'
                  contentClassName='formGroupContent'
                  style={{ marginBottom: '5px' }}
                >
                  <Label style={{ display: 'inline', marginRight: 0, fontWeight: 'bold' }}>
                    Enable Blueprint?
                    <span className='requiredStar'>*</span>
                  </Label>
                  <Switch
                    id='blueprint-enable'
                    name='blueprint-enable'
                    checked={enable}
                    label={enable ? 'Active' : 'Inactive'}
                    onChange={() => setEnableBlueprint(e => !e)}
                    style={{ marginBottom: 0, marginTop: 0 }}
                  />
                </FormGroup>
              </div>
              <div>
                <div>
                  <Label style={{ display: 'inline', marginRight: 0, marginBottom: 0 }}>
                    Next Run Date
                  </Label>
                </div>
                <div style={{ fontSize: '14px', fontWeight: 800 }}>
                  {getFieldError('Blueprint Cron') && (
                    <Icon icon='warning-sign' size={14} color={Colors.RED4} style={{ marginRight: '5px' }} />
                  )}
                  {dayjs(createCron(cronConfig === 'custom'
                    ? customCronConfig
                    : cronConfig).next().toString()).format('L LTS')} &middot;{' '}
                  <span style={{ color: Colors.GRAY3 }}>({dayjs(createCron(cronConfig === 'custom'
                    ? customCronConfig
                    : cronConfig).next().toString()).fromNow()})
                  </span>
                </div>
              </div>
            </div>
          )}

        </div>
        <div className={Classes.DIALOG_FOOTER}>
          <div className={Classes.DIALOG_FOOTER_ACTIONS}>
            <Button disabled={isSaving} onClick={() => setIsOpen(false)}>Cancel</Button>
            <Button
              disabled={isSaving || !isValidBlueprint}
              icon='cloud-upload'
              intent={Intent.PRIMARY}
              onClick={() => saveBlueprint(draftBlueprint ? draftBlueprint.id : null)}
            >
              {draftBlueprint ? 'Modify Blueprint' : 'Save Blueprint'}
            </Button>
          </div>
        </div>
      </Dialog>
    </>
  )
}

export default AddBlueprintDialog
