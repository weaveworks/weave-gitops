import styled from "styled-components";
import Flex from "./Flex";

const MessageBox = styled(Flex)`
  box-sizing: border-box;
  width: 560px;
  padding: ${({ theme }) => theme.spacing.medium}
    ${({ theme }) => theme.spacing.xl} ${({ theme }) => theme.spacing.xxl};
  border-radius: 10px;
  background-color: #ffffffd9;
  color: ${({ theme }) => theme.colors.neutral30};
`;

MessageBox.defaultProps = {
  column: true,
  shadow: true,
};

export default MessageBox;
