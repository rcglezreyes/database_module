// src/components/ScorePieChartComponent.js

import React from 'react';
import { Box, Typography } from '@mui/material';
import { PieChart, Pie, Cell, Tooltip, Legend } from 'recharts';

// Define los colores para cada segmento del pie chart
const COLORS = ['#ff9999', '#66b3ff', '#99cc99', '#ffcc99', '#c2c2f0'];

const ScorePieChartComponent = ({ data }) => {
  console.log('Datos recibidos:', data); // Verifica los datos aquí

  // Función para calcular los porcentajes
  const calculatePercentages = (data) => {
    const total = data.reduce((sum, entry) => sum + entry.student_count, 0);
    return data.map(entry => ({
      ...entry,
      studentCount: (entry.student_count / total) * 100, // Convertir a porcentaje
    }));
  };

  // Verificar si hay datos
  if (!data || data.length === 0) {
    return <Box p={2}>No hay datos disponibles.</Box>;
  }

  // Calcular los datos en porcentaje
  const dataWithPercentages = calculatePercentages(data);

  return (
    <Box p={2} sx={{ minHeight: '100%'}}>
      <PieChart width={600} height={400} sx={{ mt: 20 }}>
        <Pie
          data={dataWithPercentages}
          dataKey="studentCount"
          nameKey="range"
          outerRadius={150}
          fill="#8884d8"
          label={({ name, value }) => `${name}: ${value.toFixed(2)}%`}
        >
          {dataWithPercentages.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
          ))}
        </Pie>
        <Tooltip formatter={(value) => `${value.toFixed(2)}%`} />
        <Legend />
      </PieChart>
    </Box>
  );
};

export default ScorePieChartComponent;
