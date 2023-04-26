import React, { Dispatch, SetStateAction, useEffect } from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useGetInventory } from "../hooks/inventory";
import { Condition } from "../lib/api/core/types.pb";
import { FluxObject } from "../lib/objects";
import { ReconciledObjectsAutomation } from "./AutomationDetail";
import { filterByStatusCallback, filterConfig } from "./DataTable";
import FluxObjectsTable from "./FluxObjectsTable";
import { ReadyStatusValue, ReadyType } from "./KubeStatusIndicator";
import RequestStateHandler from "./RequestStateHandler";

interface Props {
  className?: string;
  reconciledObjectsAutomation: ReconciledObjectsAutomation;
  setCanaryStatus?: Dispatch<SetStateAction<Condition>>;
}

export const createCanaryCondition = (objs: FluxObject[]): Condition => {
  //summarize status of all canaries
  const statusTally = objs.reduce(
    (prev, current) => {
      if (!current || current.type !== "Canary") return prev;
      const condition = current.conditions[0];
      if (!condition) return prev;
      if (
        condition.reason === "Succeeded" ||
        condition.reason === "Initialized"
      )
        prev["True"] += 1;
      else if (condition.reason === "Failed") prev["False"] += 1;
      else prev["Unknown"] += 1;
      return prev;
    },
    { True: 0, False: 0, Unknown: 0 }
  );
  //create conditions object
  let condition: Condition = {
    type: ReadyType.Ready,
    status: ReadyStatusValue.None,
    message: "No canaries",
  };
  if (statusTally["False"])
    condition = {
      type: ReadyType.Ready,
      status: ReadyStatusValue.False,
      message: "One or more canaries failed",
    };
  else if (statusTally["Unknown"])
    condition = {
      type: ReadyType.Ready,
      status: ReadyStatusValue.Unknown,
      reason: "Progressing",
      message: "One or more canaries in progress",
    };
  else if (statusTally["True"])
    condition = {
      type: ReadyType.Ready,
      status: ReadyStatusValue.True,
      message: "All canaries have Succeeded",
    };
  return condition;
};

function ReconciledObjectsTable({
  className,
  reconciledObjectsAutomation,
  setCanaryStatus,
}: Props) {
  const { type, name, clusterName, namespace } = reconciledObjectsAutomation;
  const { data, isLoading, error } = useGetInventory(
    type,
    name,
    clusterName,
    namespace,
    false
  );

  //calculate canary status
  useEffect(() => {
    //if no data or in OSS don't run
    if (!data?.objects || !setCanaryStatus) return;
    const condition = createCanaryCondition(data.objects);
    return setCanaryStatus(condition);
  }, [data]);

  const initialFilterState = {
    ...filterConfig(data?.objects, "type"),
    ...filterConfig(data?.objects, "namespace"),
    ...filterConfig(data?.objects, "status", filterByStatusCallback),
  };

  const { setDetailModal } = React.useContext(AppContext);

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <FluxObjectsTable
        className={className}
        objects={data?.objects}
        onClick={setDetailModal}
        initialFilterState={initialFilterState}
      />
    </RequestStateHandler>
  );
}
export default styled(ReconciledObjectsTable).attrs({
  className: ReconciledObjectsTable.name,
})`
  td:nth-child(5),
  td:nth-child(6) {
    white-space: pre-wrap;
  }
  td:nth-child(5) {
    overflow-wrap: break-word;
    word-wrap: break-word;
  }
`;
