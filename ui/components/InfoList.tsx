import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Text from "./Text";

export type InfoField = [string, any];

const InfoList = styled(
  ({ items, className }: { className?: string; items: InfoField[] }) => {
    return (
      <table className={className}>
        <tbody>
          {_.map(items, ([k, v], index) => (
            <tr key={`item ${index}`}>
              <td>
                <Text capitalize semiBold color="neutral30">
                  {k}:
                </Text>
              </td>
              <td>{v || "-"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }
)`
  border-spacing: 0;
  tbody tr td:first-child {
    min-width: 200px;
  }
  td {
    padding: ${(props) => props.theme.spacing.xxs} 0;
    word-break: break-all;
    vertical-align: top;
    white-space: pre-wrap;
  }
  tr {
    height: 16px;
  }
`;

export default InfoList;
