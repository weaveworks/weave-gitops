import { FormControlLabel, FormGroup, Switch } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Button from "./Button";
import Flex from "./Flex";

type Props = {
  className?: string;
  loading?: boolean;
  onClick: (opts: { withSource: boolean }) => void;
};

function SyncButton({ className, loading, onClick }: Props) {
  const [withSource, setWithSource] = React.useState(true);
  return (
    <Flex align start className={className}>
      <Button
        style={{ marginRight: 8 }}
        loading={loading}
        variant="outlined"
        onClick={() => onClick({ withSource })}
      >
        Sync
      </Button>
      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              color="primary"
              checked={withSource}
              onChange={() => setWithSource(!withSource)}
            />
          }
          label="Sync with Source"
        />
      </FormGroup>
    </Flex>
  );
}

export default styled(SyncButton).attrs({ className: SyncButton.name })``;
