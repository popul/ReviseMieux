import { DarkTheme, DefaultTheme, ThemeProvider } from '@react-navigation/native';
import { useFonts } from 'expo-font';
import { Stack } from 'expo-router';
import * as SplashScreen from 'expo-splash-screen';
import { useEffect } from 'react';
import 'react-native-reanimated';

import { useColorScheme } from '@/components/useColorScheme';

export { ErrorBoundary } from 'expo-router';

export const unstable_settings = {
  initialRouteName: '(tabs)',
};

SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const [loaded, error] = useFonts({
    SpaceMono: require('../assets/fonts/SpaceMono-Regular.ttf'),
  });

  useEffect(() => {
    if (error) throw error;
  }, [error]);

  useEffect(() => {
    if (loaded) {
      SplashScreen.hideAsync();
    }
  }, [loaded]);

  if (!loaded) return null;

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
          options={{ title: 'Chapitre', headerBackTitle: 'Retour' }}
        />
        <Stack.Screen
          name="session/[id]"
          options={{ headerShown: false, gestureEnabled: false }}
        />
        <Stack.Screen
          name="capture"
          options={{ title: 'Capturer', presentation: 'fullScreenModal' }}
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
