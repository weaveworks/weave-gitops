import { CircularProgress, Tooltip } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { theme } from "..";
import { muiTheme } from "../lib/theme";
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
  const [visible, setVisible] = React.useState(false);

  React.useEffect(() => {
    loading
      ? setVisible(true)
      : setTimeout(() => {
          const date = new Date();
          setUpdated(date.toTimeString().slice(0, 8));
          setVisible(false);
        }, 1000);
  }, [loading]);

  return (
    <Flex align className={className}>
      <Tooltip title={updated} arrow placement="top">
        <p className="timestamp-p">last updated {interval} seconds ago</p>
      </Tooltip>
      <Spacer padding="base">
        <Flex>
          <CircularProgress
            size={theme.spacing.base}
            className={visible ? "in" : "out"}
          />
        </Flex>
      </Spacer>
    </Flex>
  );
}

export default styled(PollingIndicator)`
  .timestamp-p {
    color: ${muiTheme.palette.text.secondary};
    margin-top: 0;
  }
  .MuiCircularProgress-root {
    transition: opacity 2s;
    margin-bottom: 12px;
    &.in {
      opacity: 1;
    }
    &.out {
      opacity: 0;
    }
  }
`;
