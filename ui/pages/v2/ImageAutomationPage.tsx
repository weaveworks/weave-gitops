import * as React from "react";
import styled from "styled-components";
import Flex from "../../components/Flex";
import ImageAutomation from "../../components/ImageAutomation/ImageAutomation";

import Link from "../../components/Link";
import MessageBox from "../../components/MessageBox";
import Page from "../../components/Page";
import Spacer from "../../components/Spacer";
import Text from "../../components/Text";
import { useCheckCRDInstalled } from "../../hooks/imageautomation";
const crdName = "imageupdateautomations.image.toolkit.fluxcd.io";

function ImageAutomationPage() {
  const {
    data: isCRDAvailable,
    isLoading,
    error,
  } = useCheckCRDInstalled(crdName);
  return (
    <Page loading={isLoading} error={error} path={[{ label: "Image Automations" }]}>
      {!isCRDAvailable ? (
        <Flex wide tall column align>
          <MessageBox>
            <Text size="large" semiBold>
              None of the clusters you have connected in Weave GitOps have the
              requirements installed for Image Automation.
            </Text>
            <Spacer padding="medium" />
            <Text size="medium">
              To get started with this feature, follow the Flux guide to install
              the Image Reflector and Image Automation controllers on your
              cluster(s).
            </Text>
            <Spacer padding="xs" />
            <Text>
              To learn more about how to Automate container image updates to Git
              with Flux,&nbsp;
              <Link href="https://fluxcd.io/flux/guides/image-update/" newTab>
                visit our documentation
              </Link>
            </Text>
          </MessageBox>
        </Flex>
      ) : (
        <ImageAutomation />
      )}
    </Page>
  );
}

export default styled(ImageAutomationPage)``;
