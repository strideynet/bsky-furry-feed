<script lang="ts">
  import { onMount } from 'svelte';
  import { get } from 'svelte/store';

  import { goto } from '$app/navigation';
  import { agent, session } from '$lib/atp';
  import { head } from '$stores/head';

  import Button from '$components/inputs/button.svelte';
  import Field from '$components/inputs/field.svelte';

  onMount(() => head.set({ ...$head, title: 'Login' }));

  let error = '';

  const handleLogin = async (e: Event) => {
    e.preventDefault();
    error = '';
    const form = e.target as HTMLFormElement;

    const username = form.username.value as string,
      password = form.password.value as string;

    if (!username.trim() || !password.trim()) {
      error = 'Provide a valid indentifier and password';
      return;
    }

    if (!get(agent)) {
      console.error('Agent is not initialized');
      error = 'Agent is not initialized';
      return;
    }

    await get(agent)
      ?.login({
        identifier: username,
        password
      })
      .then((res) => {
        console.log({ res });

        if (!res.success) {
          console.error('Login failed', res);
          error = 'Login failed';
          return;
        }

        session.set(res.data);
        goto('/dash');
      })
      .catch((err) => {
        console.error('Login failed', err);
        error = 'Login failed';
      });
  };
</script>

<div class="mx-auto mt-8 max-w-xl">
  <form on:submit={handleLogin} class="flex flex-col items-center justify-center">
    <Field type="text" name="username" placeholder="Identifier" required />
    <Field type="password" name="password" placeholder="Password" required />
    {#if error}
      <p class="my-2 text-red-500">{error}</p>
    {/if}
    <Button type="submit">Login</Button>
  </form>
</div>
