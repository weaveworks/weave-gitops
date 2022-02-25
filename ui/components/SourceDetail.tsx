import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useListAutomations } from "../hooks/automations";
import { useListSources } from "../hooks/sources";
import { SourceRefSourceKind } from "../lib/api/core/types.pb";
import Alert from "./Alert";
import AutomationsTable from "./AutomationsTable";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import InfoList from "./InfoList";
import { computeMessage, computeReady } from "./KubeStatusIndicator";
import LoadingPage from "./LoadingPage";
import Text from "./Text";

type Props = {
  className?: string;
  type: SourceRefSourceKind;
  name: string;
  namespace: string;
  children?: JSX.Element;
  info: <T>(s: T) => { [key: string]: any };
};

function SourceDetail({ className, name, info }: Props) {
  const { data: sources, isLoading, error } = useListSources();
  const { data: automations } = useListAutomations();

  if (isLoading) {
    return <LoadingPage />;
  }

  const s = _.find(sources, { name });

  const items = info(s);

  const relevantAutomations = _.filter(automations, (a) => {
    if (!s) {
      return false;
    }

    if (a?.sourceRef?.kind == s.type && a.sourceRef.name == name) {
      return true;
    }

    return false;
  });

  const ok = computeReady(s.conditions);
  const msg = computeMessage(s.conditions);

  return (
    <div className={className}>
      <Flex align wide between>
        <div>
          <h2>{s.name}</h2>
        </div>
        <div className="page-status">
          {ok ? (
            <Icon
              color="success"
              size="medium"
              type={IconType.CheckMark}
              text={msg}
            />
          ) : (
            <Icon
              color="alert"
              size="medium"
              type={IconType.ErrorIcon}
              text={`Error: ${msg}`}
            />
          )}
        </div>
      </Flex>
      {error && (
        <Alert severity="error" title="Error" message={error.message} />
      )}
      <div>
        <h3>{s.type}</h3>
      </div>
      <div>
        <InfoList items={items} />
      </div>
      <div>
        <AutomationsTable automations={relevantAutomations} />
      </div>
    </div>
  );
}

export default styled(SourceDetail).attrs({ className: SourceDetail.name })`
  h3 {
    margin-bottom: 24px;
  }

  ${InfoList} {
    margin-bottom: 60px;
  }

  .page-status ${Icon} ${Text} {
    color: ${(props) => props.theme.colors.black};
    font-weight: normal;
  }
`;
