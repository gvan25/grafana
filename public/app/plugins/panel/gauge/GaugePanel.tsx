import React, { PureComponent } from 'react';
import { PanelProps, NullValueMode, Gauge, Themes } from '@grafana/ui';

import { getTimeSeriesVMs } from './timeSeries';
import { GaugeOptions } from './types';
import { contextSrv } from 'app/core/core';

interface Props extends PanelProps<GaugeOptions> {}

export class GaugePanel extends PureComponent<Props> {
  getTheme() {
    return contextSrv.user.lightTheme ? Themes.Light : Themes.Dark;
  }

  render() {
    const { timeSeries, width, height, onInterpolate, options } = this.props;

    const prefix = onInterpolate(options.prefix);
    const suffix = onInterpolate(options.suffix);

    const vmSeries = getTimeSeriesVMs({
      timeSeries: timeSeries,
      nullValueMode: NullValueMode.Ignore,
    });

    return (
      <Gauge
        timeSeries={vmSeries}
        {...this.props.options}
        width={width}
        height={height}
        prefix={prefix}
        suffix={suffix}
        theme={this.getTheme()}
      />
    );
  }
}
