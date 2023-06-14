import ReactMarkdown from "react-markdown";
import styled from "styled-components";

export const MarkdownEditor = styled(ReactMarkdown)`
  width: calc(100% - 24px);
  padding: ${(props) => props.theme.spacing.small};
  overflow: scroll;
  background: ${(props) => props.theme.colors.neutralGray};
  max-height: 300px;
  & a {
    color: ${(props) => props.theme.colors.primary};
  }
  ,
  & > *:first-child {
    margin-top: ${(props) => props.theme.spacing.none};
  }
  ,
  & > *:last-child {
    margin-bottom: ${(props) => props.theme.spacing.none};
  }
`;
