import { CircularProgress, Tooltip } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { theme } from "..";
import Flex from "./Flex";
import Spacer from "./Spacer";

type Props = {
  /** whether `<CircularProgress />` should be shown */
  loading: boolean;
  interval: string;
  className?: string;
};

function PollingIndicator({ className, loading, interval }: Props) {
  const date = new Date();
  const [updated, setUpdated] = React.useState(date.toTimeString().slice(0, 8));
  const [indicator, setIndicator] = React.useState(false);

  React.useEffect(() => {
    loading
      ? setIndicator(true)
      : setTimeout(() => {
          const date = new Date();
          setUpdated(date.toTimeString().slice(0, 8));
          setIndicator(false);
        }, 1000);
  }, [loading]);

  return (
    <Flex align className={className}>
      <Tooltip title={updated} arrow placement="top">
        <p>last updated {interval} seconds ago</p>
      </Tooltip>
      <Spacer padding="base">
        <CircularProgress
          size={theme.spacing.base}
          className={indicator ? "in" : "out"}
        />
      </Spacer>
    </Flex>
  );
}

export default styled(PollingIndicator)`
  .MuiCircularProgress-root {
    transition: opacity 2s;
    &.in {
      opacity: 1;
    }
    &.out {
      opacity: 0;
    }
  }
`;
