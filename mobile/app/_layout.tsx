import { DarkTheme, DefaultTheme, ThemeProvider } from '@react-navigation/native';
import { useFonts } from 'expo-font';
import { Stack, router } from 'expo-router';
import * as SplashScreen from 'expo-splash-screen';
import { useEffect, useState } from 'react';
import { Pressable, Text } from 'react-native';
import { useColorScheme } from '@/components/useColorScheme';
import { fetchDevToken, setAuthToken, getAuthToken } from '@/services/api';

export { ErrorBoundary } from 'expo-router';

export const unstable_settings = {
  initialRouteName: '(tabs)',
};

SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const [loaded, error] = useFonts({
    SpaceMono: require('../assets/fonts/SpaceMono-Regular.ttf'),
  });
  const [authReady, setAuthReady] = useState(false);

  useEffect(() => {
    if (error) throw error;
  }, [error]);

  // Fetch dev token BEFORE rendering children (prevents race condition with API calls)
  useEffect(() => {
    if (!loaded) return;

    if (getAuthToken()) {
      setAuthReady(true);
      SplashScreen.hideAsync();
      return;
    }

    fetchDevToken()
      .then((res) => {
        setAuthToken(res.token);
        console.log('Dev token set for user:', res.user_id);
      })
      .catch((err) => {
        console.warn('Failed to fetch dev token:', err.message);
      })
      .finally(() => {
        setAuthReady(true);
        SplashScreen.hideAsync();
      });
  }, [loaded]);

  if (!loaded || !authReady) return null;

  return <RootLayoutNav />;
}

function RootLayoutNav() {
  const colorScheme = useColorScheme();

  return (
    <ThemeProvider value={colorScheme === 'dark' ? DarkTheme : DefaultTheme}>
      <Stack>
        <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
        <Stack.Screen
          name="chapter/[id]"
          options={{
            title: 'Chapitre',
            headerLeft: () => (
              <Pressable testID="header-back-btn" onPress={() => router.back()}>
                <Text style={{ color: '#4A90D9', fontSize: 17 }}>‹ Retour</Text>
              </Pressable>
            ),
          }}
        />
        <Stack.Screen
          name="session/[id]"
          options={{ headerShown: false, gestureEnabled: false }}
        />
        <Stack.Screen
          name="processing"
          options={{ title: '', headerShown: false }}
        />
        <Stack.Screen
          name="onboarding"
          options={{ headerShown: false }}
        />
      </Stack>
    </ThemeProvider>
  );
}
