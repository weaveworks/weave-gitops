import { CircularProgress } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { theme } from "..";
import Flex from "./Flex";

type Props = {
  /** whether `<CircularProgress />` should be shown */
  loading: boolean;
  className?: string;
};

function PollingIndicator({ className, loading }: Props) {
  const [visible, setVisible] = React.useState(false);

  React.useEffect(() => {
    loading
      ? setVisible(true)
      : setTimeout(() => {
          setVisible(false);
        }, 1000);
  }, [loading]);

  return (
    <Flex align className={className}>
      <CircularProgress
        size={theme.spacing.base}
        className={visible ? "in" : "out"}
      />
    </Flex>
  );
}

export default styled(PollingIndicator)`
  .MuiCircularProgress-root {
    transition: opacity 2s;
    margin-bottom: 0;
    &.in {
      opacity: 1;
    }
    &.out {
      opacity: 0;
    }
  }
`;
