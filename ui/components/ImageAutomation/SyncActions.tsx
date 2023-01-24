import React from "react";
import { useSyncFluxObject } from "../../hooks/automations";
import { useToggleSuspend } from "../../hooks/flux";
import { Kind } from "../../lib/api/core/types.pb";
import Button from "../Button";
import Flex from "../Flex";
import Spacer from "../Spacer";
import SyncButton from "../SyncButton";

interface Props {
  name?: string;
  namespace?: string;
  clusterName?: string;
  kind?: Kind;
  suspended?: boolean;
}
const SyncActions = ({
  name,
  namespace,
  clusterName,
  kind,
  suspended,
}: Props) => {
  const suspend = useToggleSuspend(
    {
      objects: [
        {
          name,
          namespace,
          clusterName,
          kind,
        },
      ],
      suspend: !suspended,
    },
    "sources"
  );

  const sync = useSyncFluxObject([
    {
      name,
      namespace,
      clusterName,
      kind,
    },
  ]);
  return (
    <Flex wide start>
      <SyncButton
        onClick={() => sync.mutateAsync({ withSource: false })}
        loading={sync.isLoading}
        disabled={suspended}
        hideDropdown={true}
      />
      <Spacer padding="xs" />
      <Button onClick={() => suspend.mutateAsync()} loading={suspend.isLoading}>
        {suspended ? "Resume" : "Suspend"}
      </Button>
    </Flex>
  );
};

export default SyncActions;
