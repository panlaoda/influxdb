// Libraries
import React, {PureComponent} from 'react'
import {withRouter, WithRouterProps} from 'react-router'
import _ from 'lodash'

// Components
import {
  Button,
  ResourceCard,
  FlexBox,
  FlexDirection,
  ComponentSize,
} from '@influxdata/clockface'
import BucketContextMenu from 'src/buckets/components/BucketContextMenu'
import BucketAddDataButton from 'src/buckets/components/BucketAddDataButton'
import {FeatureFlag} from 'src/shared/utils/featureFlag'
import OverlayLink from 'src/overlays/components/OverlayLink'

// Constants
import {isSystemBucket} from 'src/buckets/constants/index'

// Types
import {Bucket} from 'src/types'
import {DataLoaderType} from 'src/types/dataLoaders'

export interface PrettyBucket extends Bucket {
  ruleString: string
}

interface Props {
  bucket: PrettyBucket
  onEditBucket: (b: PrettyBucket) => void
  onDeleteBucket: (b: PrettyBucket) => void
  onAddData: (b: PrettyBucket, d: DataLoaderType, l: string) => void
  onUpdateBucket: (b: PrettyBucket) => void
  onFilterChange: (searchTerm: string) => void
}

class BucketRow extends PureComponent<Props & WithRouterProps> {
  public render() {
    const {bucket, onDeleteBucket} = this.props
    return (
      <ResourceCard
        testID="bucket--card"
        contextMenu={
          !isSystemBucket(bucket.name) && (
            <BucketContextMenu
              bucket={bucket}
              onDeleteBucket={onDeleteBucket}
            />
          )
        }
        name={this.cardName}
        metaData={this.cardMetaItems}
      >
        {this.actionButtons}
      </ResourceCard>
    )
  }

  private get cardName(): JSX.Element {
    const {bucket} = this.props
    if (bucket.type === 'user') {
      return (
        <OverlayLink overlayID="edit-bucket" resourceID={bucket.id}>
          {onClick => (
            <ResourceCard.Name
              testID={`bucket--card ${bucket.name}`}
              onClick={onClick}
              name={bucket.name}
            />
          )}
        </OverlayLink>
      )
    }

    return (
      <ResourceCard.Name
        testID={`bucket--card ${bucket.name}`}
        name={bucket.name}
      />
    )
  }

  private get cardMetaItems(): JSX.Element[] {
    const {bucket} = this.props
    if (bucket.type === 'system') {
      return [
        <span
          className="system-bucket"
          key={`system-bucket-indicator-${bucket.id}`}
        >
          System Bucket
        </span>,
        <>Retention: {bucket.ruleString}</>,
      ]
    }

    return [<>Retention: {bucket.ruleString}</>]
  }

  private get actionButtons(): JSX.Element {
    const {bucket} = this.props
    if (bucket.type === 'user') {
      return (
        <FlexBox
          direction={FlexDirection.Row}
          margin={ComponentSize.Small}
          style={{marginTop: '4px'}}
        >
          <BucketAddDataButton
            onAddCollector={this.handleAddCollector}
            bucketID={bucket.id}
          />
          <OverlayLink overlayID="rename-bucket" resourceID={bucket.id}>
            {onClick => (
              <Button
                text="Rename"
                testID="bucket-rename"
                size={ComponentSize.ExtraSmall}
                onClick={onClick}
              />
            )}
          </OverlayLink>
          <FeatureFlag name="deleteWithPredicate">
            <OverlayLink overlayID="delete-data" resourceID={bucket.id}>
              {onClick => (
                <Button
                  text="Delete Data By Filter"
                  testID="bucket-delete-task"
                  size={ComponentSize.ExtraSmall}
                  onClick={onClick}
                />
              )}
            </OverlayLink>
          </FeatureFlag>
        </FlexBox>
      )
    }
  }

  private handleAddCollector = (): void => {
    const {
      params: {orgID},
      bucket: {id},
    } = this.props

    const link = `/orgs/${orgID}/load-data/buckets/${id}/telegrafs/new`
    this.props.onAddData(this.props.bucket, DataLoaderType.Streaming, link)
  }
}

export default withRouter<Props>(BucketRow)
