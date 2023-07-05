import {useEffect, useMemo, useState} from 'react'
import {AppBskyActorDefs, AtpAgent} from "@atproto/api";
import {
  AppBar,
  Container,
  Toolbar,
  Grid,
  Paper,
  Card,
  CardContent
} from "@mui/material";
import Typography from '@mui/material/Typography';

interface session {
  did: string
}

const useSession = (agent: AtpAgent) => {
  const [session, setSession] = useState<session | undefined>()

  useEffect(() => {
    agent.login({ identifier: import.meta.env.VITE_BSKY_USERNAME, password: import.meta.env.VITE_BSKY_PASSWORD })
      .then((res) => {
        setSession({
          did: res.data.did,
        })
      })
      .catch((err) => {
        console.error(err)
      })
  }, [agent])

  return session
}

const useGetProfile = (agent: AtpAgent, session?: session) => {
  const [profile, setProfile] = useState<AppBskyActorDefs.ProfileViewDetailed | undefined>()

  useEffect(() => {
    if (!session) {
      console.debug("no session")
      return
    }

    agent.api.app.bsky.actor.getProfile({ actor: session.did })
      .then((res) => {
        setProfile(res.data)
      })
      .catch((err) => {
        console.error(err)
      })
  }, [session])

  return profile
}

type  ProfileComponentProps = {
  profile: AppBskyActorDefs.ProfileViewDetailed,
}

const ProfileComponent = ({ profile }: ProfileComponentProps) => {
  return <>
    Name: { profile.displayName }<br/>
    Handle: { profile.handle }
  </>
}

const App = () => {
  const agent = useMemo(() => {
    return new AtpAgent({ service: 'https://bsky.social' })
  }, [])
  const session = useSession(agent)
  const profile = useGetProfile(agent, session)

  return (
    <>
      <AppBar position="static">
        <Toolbar variant="dense">
          <Typography variant="h6" color="inherit" component="div">
            Furryli.st
          </Typography>
        </Toolbar>
      </AppBar>
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Grid container spacing={3}>
          <Grid item xs={12} md={8} lg={4}>
            <Card sx={{ minWidth: 275 }}>
              <CardContent>
                <Typography variant="h4">
                  About You
                </Typography>
                <Typography variant="body1">
                  {profile ?
                    <ProfileComponent profile={profile}/>
                    : 'loading'}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={8} lg={8}>
            <Paper
              sx={{
                p: 2,
                display: 'flex',
                flexDirection: 'column',
                height: 240,
              }}
            >
            </Paper>
          </Grid>
        </Grid>
      </Container>
    </>
  )
}

export default App
