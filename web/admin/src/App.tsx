import { useContext, useEffect, useState } from "react";
import { AppBskyActorDefs, AtpAgent, AtpSessionData } from "@atproto/api";
import {
  ActionIcon,
  AppShell,
  Box,
  Group,
  Header,
  Image,
  Navbar,
  NavLink,
  rem,
  Text,
  TextInput,
  Title,
  useMantineColorScheme,
  useMantineTheme,
} from "@mantine/core";
import {
  IconHome,
  IconMoonStars,
  IconSun,
  IconTicket,
} from "@tabler/icons-react";
import { Outlet, NavLink as RouterNavLink, Navigate } from "react-router-dom";
import { AgentContext } from "./auth.tsx";

const useGetProfile = (agent: AtpAgent, session?: AtpSessionData) => {
  const [profile, setProfile] = useState<
    AppBskyActorDefs.ProfileViewDetailed | undefined
  >();

  useEffect(() => {
    if (!session) {
      return;
    }

    agent.api.app.bsky.actor
      .getProfile({ actor: session.did })
      .then((res) => {
        setProfile(res.data);
      })
      .catch((err) => {
        console.error(err);
      });
  }, [agent, session]);

  return profile;
};

export const Shell = () => {
  const { colorScheme, toggleColorScheme } = useMantineColorScheme();
  const theme = useMantineTheme();

  const agentCtx = useContext(AgentContext);
  const profile = useGetProfile(agentCtx.agent, agentCtx.session);

  if (!agentCtx.session) {
    return <Navigate to={"/login"} replace />;
  }

  return (
    <>
      <AppShell
        padding="md"
        navbar={
          <Navbar width={{ base: 300 }} p="xs">
            <Navbar.Section grow mt="xs">
              <RouterNavLink to="/">
                {({ isActive }) => (
                  <NavLink label="Home" icon={<IconHome />} active={isActive} />
                )}
              </RouterNavLink>
              <RouterNavLink to="/approval-queue">
                {({ isActive }) => (
                  <NavLink
                    label="Approval Queue"
                    icon={<IconTicket />}
                    active={isActive}
                  />
                )}
              </RouterNavLink>
            </Navbar.Section>
            <Navbar.Section>
              <Box
                sx={{
                  borderTop: `${rem(1)} solid ${
                    theme.colorScheme === "dark"
                      ? theme.colors.dark[4]
                      : theme.colors.gray[2]
                  }`,
                }}
              >
                <Title order={3}>Your User</Title>
                {profile ? (
                  <Text>
                    Handle: {profile.handle}
                    <br />
                    Username: {profile.displayName}
                  </Text>
                ) : (
                  <Text>Loading...</Text>
                )}
              </Box>
            </Navbar.Section>
          </Navbar>
        }
        header={
          <Header height={60} p="xs">
            <Group sx={{ height: "100%" }} px={20} position="apart">
              <Group>
                <Image
                  src="/dogmotif.png"
                  radius="md"
                  width={40}
                  fit="contain"
                />
                <Title>Admin</Title>
              </Group>

              <ActionIcon
                variant="default"
                onClick={() => toggleColorScheme()}
                size={30}
              >
                {colorScheme === "dark" ? (
                  <IconSun size="1rem" />
                ) : (
                  <IconMoonStars size="1rem" />
                )}
              </ActionIcon>
            </Group>
          </Header>
        }
        styles={(theme) => ({
          main: {
            backgroundColor:
              theme.colorScheme === "dark"
                ? theme.colors.dark[8]
                : theme.colors.gray[0],
          },
        })}
      >
        <Outlet />
      </AppShell>
    </>
  );
};

export const HomePage = () => {
  return <Text>Home!</Text>;
};

export const ApprovalQueuePage = () => {
  return <Text>Approval Queue!</Text>;
};

export const LoginPage = () => {
  const agentCtx = useContext(AgentContext);
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  if (agentCtx.session) {
    return <Navigate to={"/"} replace />;
  }

  const login = () => {
    agentCtx.agent.login({
      identifier: identifier,
      password: password,
    });
  };

  // LoginPage is rendered outside of Shell unlike most pages.
  return (
    <>
      <TextInput
        label="identifier"
        value={identifier}
        onChange={(evt) => setIdentifier(evt.currentTarget.value)}
      />
      <TextInput
        label="password"
        value={password}
        onChange={(evt) => setPassword(evt.currentTarget.value)}
      />
      <button onClick={login}>Log In</button>
    </>
  );
};
