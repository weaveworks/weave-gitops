import React, { Dispatch, SetStateAction, useEffect } from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useGetInventory } from "../hooks/inventory";
import { Condition } from "../lib/api/core/types.pb";
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
    //summarize status of all canaries
    let status = 0;
    data.objects.forEach((obj) => {
      if (obj.type !== "Canary") return;
      //do canaries ever have more than one condition or an undefined condition?
      let promotedCondition = obj.conditions[0];
      //succeeded < progressing < failed
      if (
        (promotedCondition.reason === "Succeeded" ||
          promotedCondition.reason === "Initialized") &&
        status < 1
      )
        return (status = 1);
      if (promotedCondition.reason === "Failed" && status < 3)
        return (status = 3);
      //I'm counting Waiting, Progressing, WaitingPromotion, Promoting, and Finalising as an "In Progress" state
      return (status = 2);
    });
    //create conditions object
    let condition: Condition = {
      type: ReadyType.Ready,
      status: ReadyStatusValue.None,
      message: "No canaries",
    };
    switch (status) {
      case 1:
        condition = {
          type: ReadyType.Ready,
          status: ReadyStatusValue.True,
          message: "All canaries have Succeeded",
        };
      case 2:
        condition = {
          type: ReadyType.Ready,
          status: ReadyStatusValue.Unknown,
          reason: "Progressing",
          message: "One or more canaries in progress",
        };
      case 3:
        condition = {
          type: ReadyType.Ready,
          status: ReadyStatusValue.False,
          message: "One or more canaries failed",
        };
    }
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
