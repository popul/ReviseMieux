import React from 'react';
import { Tabs } from 'expo-router';
import { Text } from 'react-native';

import Colors from '@/constants/Colors';
import { useColorScheme } from '@/components/useColorScheme';

function TabIcon({ emoji, color }: { emoji: string; color: string }) {
  return <Text style={{ fontSize: 22 }}>{emoji}</Text>;
}

export default function TabLayout() {
  const colorScheme = useColorScheme();

  return (
    <Tabs
      screenOptions={{
        tabBarActiveTintColor: Colors[colorScheme].tint,
        tabBarInactiveTintColor: Colors[colorScheme].tabIconDefault,
        tabBarStyle: {
          borderTopColor: Colors[colorScheme].border,
        },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          title: 'Accueil',
          tabBarIcon: ({ color }) => <TabIcon emoji="🏠" color={color} />,
          tabBarTestID: 'tab-accueil',
          headerShown: false,
        }}
      />
      <Tabs.Screen
        name="capture"
        options={{
          title: 'Capturer',
          tabBarIcon: ({ color }) => <TabIcon emoji="📸" color={color} />,
          tabBarTestID: 'tab-capturer',
          headerShown: false,
        }}
      />
      <Tabs.Screen
        name="settings"
        options={{
          title: 'Plus',
          tabBarIcon: ({ color }) => <TabIcon emoji="⚙️" color={color} />,
          tabBarTestID: 'tab-plus',
        }}
      />
    </Tabs>
  );
}
