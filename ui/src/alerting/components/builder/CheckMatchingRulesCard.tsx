// Libraries
import React, {FunctionComponent, useState, useEffect} from 'react'
import {connect} from 'react-redux'
import {get} from 'lodash'

// Components
import MatchingRuleCard from 'src/alerting/components/builder/MatchingRuleCard'
import {
  SpinnerContainer,
  TechnoSpinner,
  FlexBox,
  FlexDirection,
  AlignItems,
} from '@influxdata/clockface'

// Actions
import {getActiveTimeMachine} from 'src/timeMachine/selectors'

// API
import * as api from 'src/client'

//Types
import {NotificationRule, Check, AppState, CheckTagSet} from 'src/types'
import {EmptyState, ComponentSize, RemoteDataState} from '@influxdata/clockface'
import BuilderCard from 'src/timeMachine/components/builderCard/BuilderCard'

interface StateProps {
  check: Partial<Check>
  orgID: string
}

const CheckMatchingRulesCard: FunctionComponent<StateProps> = ({
  check,
  orgID,
}) => {
  const getMatchingRules = async (): Promise<NotificationRule[]> => {
    const tags: CheckTagSet[] = get(check, 'tags', []) || []
    const tagsList = tags
      .filter(t => t.key && t.value)
      .map(t => ['tag', `${t.key}:${t.value}`])

    // todo also: get tags from query results

    const resp = await api.getNotificationRules({
      query: [['orgID', orgID], ...tagsList] as any,
    })

    if (resp.status !== 200) {
      setMatchingRules({matchingRules: [], status: RemoteDataState.Error})
      //TODO:notify?
      return
    }

    setMatchingRules({
      matchingRules: resp.data.notificationRules,
      status: RemoteDataState.Done,
    })
  }

  const [{matchingRules, status}, setMatchingRules] = useState<{
    matchingRules: NotificationRule[]
    status: RemoteDataState
  }>({matchingRules: [], status: RemoteDataState.NotStarted})

  useEffect(() => {
    setMatchingRules({
      matchingRules,
      status: RemoteDataState.Loading,
    })
    getMatchingRules()
  }, [check.tags])

  let contents: JSX.Element

  if (
    status === RemoteDataState.NotStarted ||
    status === RemoteDataState.Loading
  ) {
    contents = (
      <SpinnerContainer spinnerComponent={<TechnoSpinner />} loading={status} />
    )
  } else if (matchingRules.length === 0) {
    contents = (
      <EmptyState
        size={ComponentSize.Small}
        className="alert-builder--card__empty"
      >
        <EmptyState.Text>
          Notification Rules configured to act on tag sets matching this Alert
          Check will show up here
        </EmptyState.Text>
        <EmptyState.Text>
          Looks like no notification rules match the tag set defined in this
          Alert Check
        </EmptyState.Text>
      </EmptyState>
    )
  } else {
    contents = (
      <>
        {matchingRules.map(r => (
          <MatchingRuleCard key={r.id} rule={r} />
        ))}
      </>
    )
  }

  return (
    <BuilderCard
      testID="builder-conditions"
      className="alert-builder--card alert-builder--conditions-card"
    >
      <BuilderCard.Header title="Matching Notification Rules" />
      <BuilderCard.Body addPadding={true} autoHideScrollbars={true}>
        <FlexBox
          direction={FlexDirection.Column}
          alignItems={AlignItems.Stretch}
          margin={ComponentSize.Medium}
        >
          {contents}
        </FlexBox>
      </BuilderCard.Body>
    </BuilderCard>
  )
}

const mstp = (state: AppState): StateProps => {
  const {
    orgs: {
      org: {id: orgID},
    },
  } = state
  const {
    alerting: {check},
  } = getActiveTimeMachine(state)

  return {check, orgID}
}

export default connect<StateProps, {}, {}>(
  mstp,
  null
)(CheckMatchingRulesCard)
