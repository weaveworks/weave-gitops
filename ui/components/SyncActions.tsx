import React from "react";
import { useSyncFluxObject } from "../hooks/automations";
import { useToggleSuspend } from "../hooks/flux";
import { Kind } from "../lib/api/core/types.pb";
import Button from "./Button";
import CustomActions from "./CustomActions";
import Flex from "./Flex";
import Spacer from "./Spacer";
import SyncButton from "./SyncButton";

interface Props {
  name?: string;
  namespace?: string;
  clusterName?: string;
  kind?: Kind;
  suspended?: boolean;
  wide?: boolean;
  hideDropdown?: boolean;
  customActions?: JSX.Element[];
}

const SyncActions = ({
  name,
  namespace,
  clusterName,
  kind,
  suspended,
  wide,
  hideDropdown,
  customActions,
}: Props) => {
  const suspend = useToggleSuspend(
    {
      objects: [
        {
          name,
          namespace,
          clusterName,
          kind: kind,
        },
      ],
      suspend: !suspended,
    },
    "object"
  );

  const sync = useSyncFluxObject([
    {
      name,
      namespace,
      clusterName,
      kind: kind,
    },
  ]);

  const syncHandler = hideDropdown
    ? () => sync.mutateAsync({ withSource: false })
    : (opts) => sync.mutateAsync(opts);

  return (
    <Flex wide={wide} start>
      <SyncButton
        onClick={syncHandler}
        loading={sync.isLoading}
        disabled={suspended}
        hideDropdown={hideDropdown}
      />
      <Spacer padding="xs" />
      <Button onClick={() => suspend.mutateAsync()} loading={suspend.isLoading}>
        {suspended ? "Resume" : "Suspend"}
      </Button>
      <CustomActions actions={customActions} />
    </Flex>
  );
};

export default SyncActions;
