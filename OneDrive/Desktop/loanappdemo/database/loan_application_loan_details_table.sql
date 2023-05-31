-- MySQL dump 10.13  Distrib 8.0.33, for Win64 (x86_64)
--
-- Host: localhost    Database: loan_application
-- ------------------------------------------------------
-- Server version	8.0.33

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `loan_details_table`
--

DROP TABLE IF EXISTS `loan_details_table`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `loan_details_table` (
  `id` int NOT NULL AUTO_INCREMENT,
  `loan_type` varchar(40) DEFAULT NULL,
  `loan_amount` float DEFAULT NULL,
  `pincode` int DEFAULT NULL,
  `tenure` int DEFAULT NULL,
  `employment_type` varchar(45) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `last_modified` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `loan_details_table`
--

LOCK TABLES `loan_details_table` WRITE;
/*!40000 ALTER TABLE `loan_details_table` DISABLE KEYS */;
INSERT INTO `loan_details_table` VALUES (1,'home',60000,452001,3,'self','2023-05-03 17:40:28','2023-05-03 17:40:28'),(8,'Personal',10000,123456,12,'Salaried','2023-05-08 16:38:33','2023-05-08 16:38:33'),(9,'Personal',10000,123456,12,'Salaried','2023-05-08 16:39:05','2023-05-08 16:39:05'),(10,'Personal',10000,123456,12,'Salaried','2023-05-08 16:42:07','2023-05-08 16:42:07'),(11,'Personal',10000,123456,12,'Salaried','2023-05-08 16:42:07','2023-05-08 16:42:07'),(12,'Personal',10000,123456,12,'Salaried','2023-05-08 16:42:50','2023-05-08 16:42:50'),(13,'Personal',10000,123456,12,'Salaried','2023-05-08 16:44:38','2023-05-08 16:44:38'),(14,'Personal',10000,123456,12,'Salaried','2023-05-08 16:44:48','2023-05-08 16:44:48'),(15,'Personal',10000,123456,12,'Salaried','2023-05-08 16:50:47','2023-05-08 16:50:47');
/*!40000 ALTER TABLE `loan_details_table` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2023-05-09 12:13:23
