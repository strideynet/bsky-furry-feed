import { useEffect, useMemo, useState } from "react";
import { AppBskyActorDefs, AtpAgent } from "@atproto/api";
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
import { Outlet, NavLink as RouterNavLink } from "react-router-dom";

interface session {
  did: string;
}

const useSession = (agent: AtpAgent) => {
  const [session, setSession] = useState<session | undefined>();

  useEffect(() => {
    agent
      .login({
        identifier: import.meta.env.VITE_BSKY_USERNAME,
        password: import.meta.env.VITE_BSKY_PASSWORD,
      })
      .then((res) => {
        setSession({
          did: res.data.did,
        });
      })
      .catch((err) => {
        console.error(err);
      });
  }, [agent]);

  return session;
};

const useGetProfile = (agent: AtpAgent, session?: session) => {
  const [profile, setProfile] = useState<
    AppBskyActorDefs.ProfileViewDetailed | undefined
  >();

  useEffect(() => {
    if (!session) {
      console.debug("no session");
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
  }, [session]);

  return profile;
};

export const Shell = () => {
  const { colorScheme, toggleColorScheme } = useMantineColorScheme();
  const theme = useMantineTheme();

  const agent = useMemo(() => {
    return new AtpAgent({ service: "https://bsky.social" });
  }, []);
  const session = useSession(agent);
  const profile = useGetProfile(agent, session);

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
  // LoginPage is rendered outside of Shell unlike most pages.
  return <h1>oooh you want to log in</h1>;
};
